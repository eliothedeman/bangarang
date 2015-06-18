package smoothie

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"github.com/gonum/plot/vg/vgsvg"
)

func testPlot(df *DataFrame, name string, mod func(*DataFrame) *DataFrame) {
	plotMulti(name, []string{"raw", "smooth"}, []*DataFrame{df, mod(df)})

}

func plotMulti(name string, names []string, frames []*DataFrame) {
	if len(names) != len(frames) {
		log.Fatal("wrong length for plots")
	}
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}

	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	lines := make([]interface{}, len(names)*2)
	x := 0
	for i := 0; i < len(lines); i += 2 {
		lines[i] = names[x]
		lines[i+1] = frames[x].PlotPoints()
		x += 1
	}

	err = plotutil.AddLinePoints(p, lines...)
	if err != nil {
		log.Fatal(err)
	}

	c := vgsvg.New(16*vg.Inch, 9*vg.Inch)

	can := draw.New(c)

	p.Draw(can)
	p.Save(16*vg.Inch/2, 9*vg.Inch/2, fmt.Sprintf("graphs/%s.png", name))
	f, err := os.Create(fmt.Sprintf("graphs/%s.svg", name))
	if err != nil {
		log.Fatal(err)
	}

	c.WriteTo(f)

}

func plotSingle(df *DataFrame, name string) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}

	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err = plotutil.AddLinePoints(p, name, df.PlotPoints())
	if err != nil {
		log.Fatal(err)
	}

	c := vgsvg.New(16*vg.Inch, 9*vg.Inch)

	can := draw.New(c)

	p.Draw(can)
	p.Save(16*vg.Inch/2, 9*vg.Inch/2, fmt.Sprintf("graphs/%s.png", name))
	f, err := os.Create(fmt.Sprintf("graphs/%s.svg", name))
	if err != nil {
		log.Fatal(err)
	}

	c.WriteTo(f)

}

func randDF(size int) *DataFrame {
	df := NewSignal(200, rand.Float64()*15)
	df = df.Add(NewSignal(200, rand.Float64()*7))
	return df.Add(Noise(200))
}

type mod func(df *DataFrame) *DataFrame

var (
	test_mods = map[string]mod{
		"moving_average": func(df *DataFrame) *DataFrame {
			return df.MovingAverage(10)
		},
		"weighted_average": func(df *DataFrame) *DataFrame {
			return df.WeightedMovingAverage(10, LinearWeighting)
		},
		"double_smooth": func(df *DataFrame) *DataFrame {
			return df.DoubleExponentialSmooth(0.2, 0.3)
		},
		"single_smooth": func(df *DataFrame) *DataFrame {
			return df.SingleExponentialSmooth(0.3)
		},
	}
)

func TestPlotDF(t *testing.T) {
	rand := randDF(200)

	for k, v := range test_mods {
		testPlot(rand, k, v)
	}
}
