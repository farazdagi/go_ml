/*
Linear Regression implementation
*/
package ml

import (
	"github.com/alonsovidales/matrix"
	"fmt"
	"strconv"
	"math"
	"math/rand"
	"time"
	"io/ioutil"
	"strings"
)

type LinReg struct {
	X [][]float64 // Training set of values for each feature, the first dimension are the test cases
	Y []float64 // The training set with values to be predicted
	// 1st dim -> layer, 2nd dim -> neuron, 3rd dim theta
	Theta []float64
}

func (lr *LinReg) rollThetasGrad(x [][][]float64) [][]float64 {
	return x[0]
}

func (lr *LinReg) unrollThetasGrad(x [][]float64) [][][]float64 {
	return [][][]float64{
		x,
	}
}

func (lr *LinReg) setTheta(t [][][]float64) {
	lr.Theta = t[0][0]
}

func (lr *LinReg) getTheta() [][][]float64 {
	return [][][]float64{
		[][]float64{
			lr.Theta,
		},
	}
}

func (lr *LinReg) CostFunction(lambda float64, calcGrad bool) (j float64, grad [][][]float64, err error) {
	if len(lr.Y) != len(lr.X) {
		err = fmt.Errorf(
			"The number of test cases (X) %d doesn't corresponds with the number of values (Y) %d",
			len(lr.X),
			len(lr.Y))
		return
	}

	if len(lr.Theta) != len(lr.X[0]) {
		err = fmt.Errorf(
			"The Theta arg has a lenght of %d and the input data %d",
			len(lr.Theta),
			len(lr.X[0]))
		return
	}

	auxTheta := make([]float64, len(lr.Theta))
	copy(auxTheta, lr.Theta)
	theta := [][]float64{auxTheta}

	m := float64(len(lr.X))
	y := [][]float64{lr.Y}

	pred := mt.Trans(mt.Mult(lr.X, mt.Trans(theta)))
	errors := mt.SumAll(mt.Apply(mt.Sub(pred, y), powTwo)) / (2 * m)
	regTerm := (lambda / (2 * m)) * mt.SumAll(mt.Apply([][]float64{lr.Theta[1:]}, powTwo))

	j = errors + regTerm
	theta[0][0] = 0
	grad = [][][]float64{mt.Sum(mt.MultBy(mt.Mult(mt.Sub(pred, y), lr.X), 1 / m), mt.MultBy(theta, lambda / m))}

	return
}

func (lr *LinReg) InitializeTheta() {
	lr.Theta = make([]float64, len(lr.X[0]))
}

// Loads information from the local file located at filePath, and after parse
// it, returns the LinReg ready to be used with all the information loaded
// The file format is:
//      X11 X12 ... X1N Y1
//      X21 X22 ... X2N Y2
//      ... ... ... ... ..
//      XN1 XN2 ... XNN YN
//
// Note: Use a single space as separator
func LoadFile(filePath string) (data *LinReg) {
	strInfo, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	data = new(LinReg)

	trainingData := strings.Split(string(strInfo), "\n")
	for _, line := range trainingData {
		if line == "" {
			break
		}

		var values []float64
		for _, value := range strings.Split(line, " ") {
			floatVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				panic(err)
			}
			values = append(values, floatVal)
		}
		data.X = append(data.X, values[:len(values)-1])
		data.Y = append(data.Y, values[len(values)-1])
	}

	return
}

func (data *LinReg) shuffle() (shuffledData *LinReg) {
	rand.Seed(int64(time.Now().Nanosecond()))

	shuffledData = &LinReg{
		X: make([][]float64, len(data.X)),
		Y: make([]float64, len(data.Y)),
	}

	for i, v := range rand.Perm(len(data.X)) {
		shuffledData.X[v] = data.X[i]
		shuffledData.Y[v] = data.Y[i]
	}

	shuffledData.Theta = data.Theta

	return
}

func (data *LinReg) Hipotesis(x []float64) (r float64) {
	for i := 0; i < len(x); i++ {
		r += x[i] * data.Theta[i]
	}

	return
}

// This metod splits the given data in three sets: training, cross validation,
// test. In order to calculate the optimal theta, tries with different
// possibilities and the training data, and check the best match with the cross
// validations, after obtain the best lambda, check the perfomand against the
// test set of data
func (data *LinReg) MinimizeCost(maxIters int, suffleData bool, verbose bool) (finalCost float64, trainingData *LinReg, testData *LinReg) {
	lambdas := []float64{3000, 0.0, 0.001, 0.003, 0.01, 0.03, 0.1, 0.3, 1, 3, 10, 30, 100, 300}

	if suffleData {
		data = data.shuffle()
	}

	// Get the 60% of the data as training data, 20% as cross validation, and
	// the remaining 20% as test data
	training := int64(float64(len(data.X)) * 0.6)
	cv := int64(float64(len(data.X)) * 0.8)

	trainingData = &LinReg{
		X: data.X[:training],
		Y: data.Y[:training],
		Theta: data.Theta,
	}

	cvData := &LinReg{
		X: data.X[training:cv],
		Y: data.Y[training:cv],
		Theta: data.Theta,
	}
	testData = &LinReg{
		X: data.X[cv:],
		Y: data.Y[cv:],
		Theta: data.Theta,
	}

	// Launch a process for each lambda in order to obtain the one with best
	// performance
	bestJ := math.Inf(1)
	bestLambda := 0.0
	initTheta := make([]float64, len(trainingData.Theta))
	copy(initTheta, trainingData.Theta)

	for _, posLambda := range lambdas {
		if verbose {
			fmt.Println("Checking Lambda:", posLambda)
		}
		copy(trainingData.Theta, initTheta)
		Fmincg(trainingData, posLambda, 10, verbose)
		cvData.Theta = trainingData.Theta

		j, _, _ := cvData.CostFunction(posLambda, false)

		if bestJ > j {
			bestJ = j
			bestLambda = posLambda
		}
	}

	// Include the cross validation cases into the training for the final train
	trainingData.X = append(trainingData.X, cvData.X...)
	trainingData.Y = append(trainingData.Y, cvData.Y...)

	if verbose {
		fmt.Println("Lambda:", bestLambda)
		fmt.Println("Training with the 80% of the samples...")
	}
	Fmincg(trainingData, bestLambda, maxIters, verbose)

	testData.Theta = trainingData.Theta
	data.Theta = trainingData.Theta

	finalCost, _, _ = testData.CostFunction(bestLambda, false)

	return
}
