package graphite_grafana_annotation

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/event"
)
import "github.com/marpaia/graphite-golang"

const (
	ANNOTATION_PREFIX = "bangarang.annotation"
)

func init() {
	escalation.LoadFactory("grafana_graphite_annotation", NewGrafanaGraphite)
}

type GrafanaGraphiteAnnotation struct {
	client *graphite.Graphite
}

type GrafanaGraphiteAnnotationConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// bangarang.annotation.{status}.{host}.{service}
func formatName(i *event.Incident) string {
	return strings.Replace(fmt.Sprintf("%s.%s.%s.%s", ANNOTATION_PREFIX, event.Status(i.Status), i.Tags.Get("host"), strings.Replace(i.FormatDescription(), ".", "_", -1)), " ", "_", -1)
}

func (g *GrafanaGraphiteAnnotation) Send(i *event.Incident) error {
	err := g.client.SendMetric(graphite.NewMetric(formatName(i), fmt.Sprintf("%f", i.Metric), time.Now().Unix()))
	return err
}

func (g *GrafanaGraphiteAnnotation) ConfigStruct() interface{} {
	return &GrafanaGraphiteAnnotationConfig{}
}

func (g *GrafanaGraphiteAnnotation) Init(i interface{}) error {
	logrus.Info("Initializing Grafana Graphite Annotation")
	c, ok := i.(*GrafanaGraphiteAnnotationConfig)
	if !ok {
		return fmt.Errorf("Incorrect config type. Expecting GrafanaGraphiteAnnotationConfig not %+v", i)
	}

	client, err := graphite.NewGraphite(c.Host, c.Port)
	if err != nil {
		return err
	}
	g.client = client
	return nil
}

func NewGrafanaGraphite() escalation.Escalation {
	return &GrafanaGraphiteAnnotation{}
}
