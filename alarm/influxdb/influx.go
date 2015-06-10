package influxdb

import (
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
	"github.com/influxdb/influxdb/client"
)

const (
	BATCH_SIZE = 10
)

func init() {
	alarm.LoadFactory("influxdb", NewInflux)
	alarm.LoadFactory("influx", NewInflux)
}

type Influx struct {
	client                    *client.Client
	points                    []client.Point
	index                     int
	database, retentionPolicy string
}

type InfluxConfig struct {
	Host            string `json:"host"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	RetentionPolicy string `json:"retention_policy"`
	Database        string `json:"database"`
}

func (c *Influx) Send(i *event.Incident) error {
	e := i.Event
	c.points[c.index] = client.Point{
		Name: e.Service,
		Tags: e.Tags,
		Fields: map[string]interface{}{
			"value":       e.Metric,
			"host":        e.Host,
			"service":     e.Service,
			"sub_service": e.SubService,
		},
		Timestamp: time.Now(),
	}

	c.index += 1

	if c.index%BATCH_SIZE == 0 {
		logrus.Debugf("Writing %d events to influxdb db:%s", BATCH_SIZE, c.database)
		c.index = 0
		_, err := c.client.Write(client.BatchPoints{
			Database:        c.database,
			RetentionPolicy: c.retentionPolicy,
			Points:          c.points,
		})
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func (c *Influx) ConfigStruct() interface{} {
	return &InfluxConfig{}
}

func (c *Influx) Init(i interface{}) error {
	logrus.Info("Initializing Influx client.")

	conf := i.(*InfluxConfig)
	cli, err := client.NewClient(client.Config{
		URL: url.URL{
			Scheme: "http",
			Host:   conf.Host,
		},
		Username: conf.Username,
		Password: conf.Password,
	})
	c.client = cli
	return err
}

func NewInflux() alarm.Alarm {
	return &Influx{
		points: make([]client.Point, BATCH_SIZE),
	}
}
