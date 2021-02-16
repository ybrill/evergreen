package model

import (
	"testing"

	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/utility"
	"github.com/stretchr/testify/assert"
)

func TestSubscriberModelsGithubStatusAPI(t *testing.T) {
	assert := assert.New(t)

	target := event.GithubPullRequestSubscriber{
		Owner:    "me",
		Repo:     "mine",
		PRNumber: 5,
		Ref:      "abc",
	}
	prSubscriber := event.Subscriber{
		Type:   event.GithubPullRequestSubscriberType,
		Target: &target,
	}
	apiPrSubscriber := APISubscriber{}
	err := apiPrSubscriber.BuildFromService(prSubscriber)
	assert.NoError(err)

	origPrSubscriberInterface, err := apiPrSubscriber.ToService()
	assert.NoError(err)

	origPrSubscriber, ok := origPrSubscriberInterface.(event.Subscriber)
	assert.True(ok)
	assert.EqualValues(prSubscriber.Type, origPrSubscriber.Type)
	assert.EqualValues(target, origPrSubscriber.Target)

	// incoming subscribers have target serialized as a map
	incoming := APISubscriber{
		Type: utility.ToStringPtr(event.GithubPullRequestSubscriberType),
		Target: map[string]interface{}{
			"owner":     "me",
			"repo":      "mine",
			"pr_number": 5,
			"ref":       "abc",
		},
	}

	serviceModel, err := incoming.ToService()
	assert.NoError(err)
	assert.EqualValues(origPrSubscriber, serviceModel)
}

func TestSubscriberModelsWebhook(t *testing.T) {
	assert := assert.New(t)

	target := event.WebhookSubscriber{
		URL:     "foo",
		Secret:  []byte("bar"),
		Headers: []event.WebhookHeader{},
	}
	webhookSubscriber := event.Subscriber{
		Type:   event.EvergreenWebhookSubscriberType,
		Target: &target,
	}
	apiWebhookSubscriber := APISubscriber{}
	err := apiWebhookSubscriber.BuildFromService(webhookSubscriber)
	assert.NoError(err)

	origWebhookSubscriberInterface, err := apiWebhookSubscriber.ToService()
	assert.NoError(err)

	origWebhookSubscriber, ok := origWebhookSubscriberInterface.(event.Subscriber)
	assert.True(ok)
	assert.EqualValues(webhookSubscriber.Type, origWebhookSubscriber.Type)
	assert.EqualValues(target, origWebhookSubscriber.Target)

	// incoming subscribers have target serialized as a map
	incoming := APISubscriber{
		Type: utility.ToStringPtr(event.EvergreenWebhookSubscriberType),
		Target: map[string]interface{}{
			"url":    "foo",
			"secret": "bar",
		},
	}

	serviceModel, err := incoming.ToService()
	assert.NoError(err)
	assert.EqualValues(origWebhookSubscriber, serviceModel)
}

func TestSubscriberModelsJIRAIssue(t *testing.T) {
	assert := assert.New(t)

	target := event.JIRAIssueSubscriber{
		Project:   "ABC",
		IssueType: "123",
	}
	jiraIssueSubscriber := event.Subscriber{
		Type:   event.JIRAIssueSubscriberType,
		Target: &target,
	}
	apiJIRAIssueSubscriber := APISubscriber{}
	err := apiJIRAIssueSubscriber.BuildFromService(jiraIssueSubscriber)
	assert.NoError(err)

	origJIRAIssueSubscriberInterface, err := apiJIRAIssueSubscriber.ToService()
	assert.NoError(err)

	origJIRAIssueSubscriber, ok := origJIRAIssueSubscriberInterface.(event.Subscriber)
	assert.True(ok)
	assert.EqualValues(jiraIssueSubscriber.Type, origJIRAIssueSubscriber.Type)
	assert.EqualValues(target, origJIRAIssueSubscriber.Target)

	// incoming subscribers have target serialized as a map
	incoming := APISubscriber{
		Type: utility.ToStringPtr(event.JIRAIssueSubscriberType),
		Target: map[string]interface{}{
			"project":    "ABC",
			"issue_type": "123",
		},
	}

	serviceModel, err := incoming.ToService()
	assert.NoError(err)
	assert.EqualValues(origJIRAIssueSubscriber, serviceModel)
}

func TestSubscriberModelsSlack(t *testing.T) {
	assert := assert.New(t)

	slackSubscriber := event.Subscriber{
		Type:   event.SlackSubscriberType,
		Target: "slack message",
	}

	apiSlackSubscriber := APISubscriber{}
	err := apiSlackSubscriber.BuildFromService(slackSubscriber)
	assert.NoError(err)

	origSlackSubscriber, err := apiSlackSubscriber.ToService()
	assert.NoError(err)
	assert.EqualValues(slackSubscriber, origSlackSubscriber)
}
