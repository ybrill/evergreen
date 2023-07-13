package evergreen

import (
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DefaultBufferIntervalSeconds   = 60
	DefaultBufferTargetPerInterval = 20
)

// NotifyConfig hold logging and email settings for the notify package.
type NotifyConfig struct {
	BufferTargetPerInterval int       `bson:"buffer_target_per_interval" json:"buffer_target_per_interval" yaml:"buffer_target_per_interval"`
	BufferIntervalSeconds   int       `bson:"buffer_interval_seconds" json:"buffer_interval_seconds" yaml:"buffer_interval_seconds"`
	SES                     SESConfig `bson:"ses" json:"ses" yaml:"ses"`
}

func (c *NotifyConfig) SectionId() string { return "notify" }

func (c *NotifyConfig) Get(env Environment) error {
	ctx, cancel := env.Context()
	defer cancel()
	coll := env.DB().Collection(ConfigCollection)

	res := coll.FindOne(ctx, byId(c.SectionId()))
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			*c = NotifyConfig{}
			return nil
		}

		return errors.Wrapf(err, "getting config section '%s'", c.SectionId())
	}

	if err := res.Decode(c); err != nil {
		return errors.Wrapf(err, "decoding config section '%s'", c.SectionId())
	}

	return nil
}

func (c *NotifyConfig) Set() error {
	env := GetEnvironment()
	ctx, cancel := env.Context()
	defer cancel()
	coll := env.DB().Collection(ConfigCollection)

	_, err := coll.ReplaceOne(ctx, byId(c.SectionId()), c, options.Replace().SetUpsert(true))
	return errors.Wrapf(err, "updating section '%s'", c.SectionId())
}

func (c *NotifyConfig) ValidateAndDefault() error {
	if c.BufferIntervalSeconds <= 0 {
		c.BufferIntervalSeconds = DefaultBufferIntervalSeconds
	}
	if c.BufferTargetPerInterval <= 0 {
		c.BufferTargetPerInterval = DefaultBufferTargetPerInterval
	}

	// Cap to 100 jobs/sec per server.
	jobsPerSecond := c.BufferIntervalSeconds / c.BufferTargetPerInterval
	if jobsPerSecond > maxNotificationsPerSecond {
		return errors.Errorf("maximum notification jobs per second is %d", maxNotificationsPerSecond)

	}

	return nil
}

// SESConfig configures the SES email sender.
type SESConfig struct {
	// SenderAddress is the default address that emails are sent from.
	SenderAddress string `bson:"sender_address" json:"sender_address" yaml:"sender_address"`
	// AWSKey is the key to use when connecting with AWS.
	AWSKey string `bson:"aws_key" json:"aws_key" yaml:"aws_key"`
	// AWSSecret is the secret to use when connecting with AWS.
	AWSSecret string `bson:"aws_secret" json:"aws_secret" yaml:"aws_secret"`
	// AWSRegion is the region to use when connecting with AWS.
	AWSRegion string `bson:"aws_region" json:"aws_region" yaml:"aws_region"`
}

func (c *SESConfig) validate() error {
	catcher := grip.NewBasicCatcher()

	if c.SenderAddress == "" {
		catcher.New("must provide sender address")
	}
	if c.AWSKey == "" {
		catcher.New("must provide AWS key")
	}
	if c.AWSSecret == "" {
		catcher.New("must provide AWS secret")
	}

	if c.AWSRegion == "" {
		c.AWSRegion = DefaultEC2Region
	}

	return catcher.Resolve()
}
