// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	influxdata "github.com/influxdata/influxdb/client/v2"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/transformers/senml"
	writer "github.com/mainflux/mainflux/writers/influxdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const valueFields = 5

var (
	port        string
	testLog, _  = log.New(os.Stdout, log.Info.String())
	testDB      = "test"
	streamsSize = 250
	selectMsgs  = "SELECT * FROM test..messages"
	dropMsgs    = "DROP SERIES FROM messages"
	client      influxdata.Client
	clientCfg   = influxdata.HTTPConfig{
		Username: "test",
		Password: "test",
	}
	subtopic = "topic"
)

var (
	v       float64 = 5
	stringV         = "value"
	boolV           = true
	dataV           = "base64"
	sum     float64 = 42
)

// This is utility function to query the database.
func queryDB(cmd string) ([][]interface{}, error) {
	q := influxdata.Query{
		Command:  cmd,
		Database: testDB,
	}
	response, err := client.Query(q)
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	if len(response.Results[0].Series) == 0 {
		return nil, nil
	}
	// There is only one query, so only one result and
	// all data are stored in the same series.
	return response.Results[0].Series[0].Values, nil
}

func TestSave(t *testing.T) {
	repo := writer.New(client, testDB)

	cases := []struct {
		desc         string
		msgsNum      int
		expectedSize int
	}{
		{
			desc:         "save a single message",
			msgsNum:      1,
			expectedSize: 1,
		},
		{
			desc:         "save a batch of messages",
			msgsNum:      streamsSize,
			expectedSize: streamsSize,
		},
	}

	for _, tc := range cases {
		// Clean previously saved messages.
		_, err := queryDB(dropMsgs)
		require.Nil(t, err, fmt.Sprintf("Cleaning data from InfluxDB expected to succeed: %s.\n", err))

		now := time.Now().Unix()
		msg := senml.Message{
			Channel:    "45",
			Publisher:  "2580",
			Protocol:   "http",
			Name:       "test name",
			Unit:       "km",
			UpdateTime: 5456565466,
		}
		var msgs []senml.Message

		for i := 0; i < tc.msgsNum; i++ {
			// Mix possible values as well as value sum.
			count := i % valueFields
			switch count {
			case 0:
				msg.Subtopic = subtopic
				msg.Value = &v
			case 1:
				msg.BoolValue = &boolV
			case 2:
				msg.StringValue = &stringV
			case 3:
				msg.DataValue = &dataV
			case 4:
				msg.Sum = &sum
			}

			msg.Time = float64(now + int64(i))
			msgs = append(msgs, msg)
		}

		err = repo.Save(msgs...)
		assert.Nil(t, err, fmt.Sprintf("Save operation expected to succeed: %s.\n", err))

		row, err := queryDB(selectMsgs)
		assert.Nil(t, err, fmt.Sprintf("Querying InfluxDB to retrieve data expected to succeed: %s.\n", err))

		count := len(row)
		assert.Equal(t, tc.expectedSize, count, fmt.Sprintf("Expected to have %d messages saved, found %d instead.\n", tc.expectedSize, count))
	}
}
