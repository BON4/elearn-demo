package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/BON4/elearn-demo/e2e/helpers"
	"github.com/stretchr/testify/assert"
)

func Test_Create_And_Publish_Course(t *testing.T) {
	suite := SetupSuite(t)
	defer suite.TearDown(t)

	db := helpers.OpenDB(suite.DBUrl)
	defer db.Close()

	helpers.CleanupDB(db)

	createResp, err := helpers.DoPost(suite.AppURL+"/courses", map[string]any{
		"title":       "Test course",
		"description": "This is test description 2",
		"author_id":   "579f3546-cb4b-4ddd-8991-3be8d4bf8a91",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created map[string]any
	json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	assert.True(t, helpers.CourseExists(db, id))

	// Turn off worker
	swithResp, err := helpers.DoPost(suite.AppURL+"/tests/pause-worker", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, swithResp.StatusCode)

	publishResp, err := helpers.DoPost(suite.AppURL+"/courses/"+id+"/publish", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, publishResp.StatusCode)

	var status string
	db.QueryRow("SELECT status FROM courses WHERE id=$1", id).Scan(&status)

	assert.Equal(t, "published", status)

	// Check if outbox event with status "pending" is added to postgres
	assert.True(t, helpers.OutboxEventExists(db, id, "CoursePublished"))
	eventStatus, err := helpers.GetOutboxEventStatus(db, id, "CoursePublished")
	assert.NoError(t, err)
	assert.Equal(t, "pending", eventStatus)

	// Turn on worker
	swithResp, err = helpers.DoPost(suite.AppURL+"/tests/resume-worker", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, swithResp.StatusCode)

	// Wait for worker
	time.Sleep(time.Second)

	// Check if outbox event with status "processed" is added to postgres
	// and verify that there is only one event (status changed from pending to processed)
	eventStatus, err = helpers.GetOutboxEventStatus(db, id, "CoursePublished")
	assert.NoError(t, err)
	assert.Equal(t, "processed", eventStatus)
	assert.Equal(t, 1, helpers.GetOutboxEventCount(db, id, "CoursePublished"))
}
