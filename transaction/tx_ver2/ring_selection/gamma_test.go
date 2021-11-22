package ring_selection

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestNewGammaPicker(t *testing.T) {
	shape := 0.6168096414830899
	scale := 34565.54623186987
	gp := NewGammaPicker(gammaParam{
		shape: shape,
		scale: scale,
	})

	mean := shape * scale
	variant := shape * scale * scale
	fmt.Println(mean, variant)

	for i := 0; i < 1000; i++ {
		s := gp.Rand()
		fmt.Println(i, s)
	}
}

func TestDrawSample(t *testing.T) {
	isPRV := false
	gp := NewGammaPicker(gammaParam{
		shape: prvGammaShape,
		scale: prvGammaScale,
	})
	if !isPRV {
		gp = NewGammaPicker(gammaParam{
			shape: tokenGammaShape,
			scale: tokenGammaScale,
		})
	}

	numSamples := 60000
	sampleCount := 0

	dataFileName := "prv.csv"
	if !isPRV {
		dataFileName = "token.csv"
	}
	dataFile, err := os.Create(dataFileName)
	if err != nil {
		panic(err)
	}
	defer dataFile.Close()
	dataWriter := csv.NewWriter(dataFile)


	records := make([][]string, 0)
	records = append(records, []string{"aged"})
	for sampleCount < numSamples {
		if sampleCount % 1000 == 0 {
			fmt.Printf("sampleCount: %v\n", sampleCount)
		}
		x := gp.Rand()
		aged := uint64(math.Round(x * unitTime / 40))
		records = append(records, []string{fmt.Sprintf("%v", aged)})
		sampleCount++
	}

	err = dataWriter.WriteAll(records)
	if err != nil {
		panic(err)
		return
	}

}
