package lr

import (
	"backend/utils"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LinearRegression struct {
	Weights []float64
	Bias    float64
	TrainingLoss []float64
	Converged    bool
	xMeans []float64
	xStds  []float64
	yMean  float64
	yStd   float64
}

func New(numFeatures int) *LinearRegression {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	weights := make([]float64, numFeatures)
	for i := range weights {
		weights[i] = (rng.Float64() - 0.5) * 0.02
	}
	return &LinearRegression{
		Weights:      weights,
		Bias:         0.0,
		TrainingLoss: make([]float64, 0),
		Converged:    false,
		xMeans:       make([]float64, numFeatures),
		xStds:        make([]float64, numFeatures),
	}
}

func (lr *LinearRegression) Predict(x []float64) float64 {
	if len(x) != len(lr.Weights) {
		return 0.0
	}
	result := lr.Bias
	for i, weight := range lr.Weights {
		result += weight * x[i]
	}
	return result
}

func (lr *LinearRegression) PredictBatch(xs [][]float64) []float64 {
	predictions := make([]float64, len(xs))
	for i, x := range xs {
		predictions[i] = lr.Predict(x)
	}
	return predictions
}

func (lr *LinearRegression) Fit(xs [][]float64, ys []float64, epochs int, lrRate float64, workers int) error {
	if len(xs) != len(ys) {
		return fmt.Errorf("xs y ys deben tener la misma longitud")
	}
	if len(xs) == 0 {
		return fmt.Errorf("los datos de entrenamiento no pueden estar vacíos")
	}
	if len(xs[0]) != len(lr.Weights) {
		return fmt.Errorf("dimensión de características no coincide")
	}
	n := len(xs)
	numFeatures := len(lr.Weights)
	lr.TrainingLoss = make([]float64, 0, epochs)
	lr.xMeans = make([]float64, numFeatures)
	lr.xStds = make([]float64, numFeatures)
	for j := 0; j < numFeatures; j++ {
		var sum, sumSq float64
		for i := 0; i < n; i++ {
			sum += xs[i][j]
			sumSq += xs[i][j] * xs[i][j]
		}
		lr.xMeans[j] = sum / float64(n)
		variance := (sumSq / float64(n)) - (lr.xMeans[j] * lr.xMeans[j])
		lr.xStds[j] = math.Sqrt(math.Max(variance, 1e-8))
	}
	ySum := 0.0
	for _, y := range ys {
		ySum += y
	}
	lr.yMean = ySum / float64(n)
	yVar := 0.0
	for _, y := range ys {
		yVar += (y - lr.yMean) * (y - lr.yMean)
	}
	lr.yStd = math.Sqrt(yVar / float64(n))
	xsNorm := make([][]float64, n)
	ysNorm := make([]float64, n)
	for i := 0; i < n; i++ {
		xsNorm[i] = make([]float64, numFeatures)
		for j := 0; j < numFeatures; j++ {
			if lr.xStds[j] > 1e-8 {
				xsNorm[i][j] = (xs[i][j] - lr.xMeans[j]) / lr.xStds[j]
			} else {
				xsNorm[i][j] = 0
			}
		}
		if lr.yStd > 1e-8 {
			ysNorm[i] = (ys[i] - lr.yMean) / lr.yStd
		} else {
			ysNorm[i] = 0
		}
	}
	currentLR := lrRate
	bestLoss := math.Inf(1)
	patienceCounter := 0
	maxPatience := 15
	fmt.Printf("Iniciando entrenamiento multivariable con %d características\n", numFeatures)
	fmt.Printf("Pesos iniciales: ")
	for i, w := range lr.Weights {
		fmt.Printf("W%d=%.4f ", i, w)
	}
	fmt.Printf("Sesgo=%.4f\n", lr.Bias)
	
	for epoch := 0; epoch < epochs; epoch++ {
		var wg sync.WaitGroup
		gradW := make([]chan float64, numFeatures)
		gradB := make(chan float64, workers)
		for j := 0; j < numFeatures; j++ {
			gradW[j] = make(chan float64, workers)
		}
		batchSize := max(1, n/workers)
		for w := 0; w < workers; w++ {
			start := w * batchSize
			end := utils.Min((w+1)*batchSize, n)
			if start >= n {
				break
			}
			wg.Add(1)
			go func(startIdx, endIdx int) {
				defer wg.Done()
				dw := make([]float64, numFeatures)
				var db float64
				batchLen := endIdx - startIdx
				for i := startIdx; i < endIdx; i++ {
					yPred := lr.Bias
					for j := 0; j < numFeatures; j++ {
						yPred += lr.Weights[j] * xsNorm[i][j]
					}
					err := yPred - ysNorm[i]
					for j := 0; j < numFeatures; j++ {
						dw[j] += err * xsNorm[i][j]
					}
					db += err
				}
				for j := 0; j < numFeatures; j++ {
					gradW[j] <- dw[j] / float64(batchLen)
				}
				gradB <- db / float64(batchLen)
			}(start, end)
		}
		wg.Wait()
		for j := 0; j < numFeatures; j++ {
			close(gradW[j])
		}
		close(gradB)
		totalDW := make([]float64, numFeatures)
		var totalDB float64
		var gradCount float64
		for j := 0; j < numFeatures; j++ {
			for v := range gradW[j] {
				totalDW[j] += v
				if j == 0 {
					gradCount++
				}
			}
		}
		for v := range gradB {
			totalDB += v
		}
		if gradCount > 0 {
			for j := 0; j < numFeatures; j++ {
				totalDW[j] /= gradCount
			}
			totalDB /= gradCount
		}
		const maxGrad = 1.0
		for j := 0; j < numFeatures; j++ {
			totalDW[j] = math.Max(-maxGrad, math.Min(maxGrad, totalDW[j]))
		}
		totalDB = math.Max(-maxGrad, math.Min(maxGrad, totalDB))
		
		prevWeights := make([]float64, numFeatures)
		copy(prevWeights, lr.Weights)
		prevBias := lr.Bias
		
		for j := 0; j < numFeatures; j++ {
			lr.Weights[j] -= currentLR * totalDW[j]
		}
		lr.Bias -= currentLR * totalDB
		
		if epoch%100 == 0 || epoch < 5 || epoch == epochs-1 {
			loss := lr.calculateLossMultivariate(xsNorm, ysNorm)
			fmt.Printf("Época %4d: Pérdida=%.6f, LR=%.6f\n", epoch, loss, currentLR)
			fmt.Printf("    Pesos: ")
			for j := 0; j < numFeatures; j++ {
				change := lr.Weights[j] - prevWeights[j]
				fmt.Printf("W%d=%.4f(Δ%+.4f) ", j, lr.Weights[j], change)
			}
			biasChange := lr.Bias - prevBias
			fmt.Printf("Sesgo=%.4f(Δ%+.4f)\n", lr.Bias, biasChange)
		}
		
		if epoch%20 == 0 || epoch == epochs-1 {
			loss := lr.calculateLossMultivariate(xsNorm, ysNorm)
			lr.TrainingLoss = append(lr.TrainingLoss, loss)
			if loss < bestLoss {
				bestLoss = loss
				patienceCounter = 0
			} else {
				patienceCounter++
				if patienceCounter >= maxPatience {
					currentLR *= 0.7
					patienceCounter = 0
					fmt.Printf("    → Tasa de aprendizaje reducida a %.6f (paciencia excedida)\n", currentLR)
					if currentLR < 1e-8 {
						lr.Converged = true
						fmt.Printf("    → Convergió: Tasa de aprendizaje demasiado pequeña\n")
						break
					}
				}
			}
			if loss < 1e-6 {
				lr.Converged = true
				fmt.Printf("    → Convergió: Umbral de pérdida alcanzado\n")
				break
			}
		}
	}
	if lr.yStd > 1e-8 {
		for j := 0; j < numFeatures; j++ {
			if lr.xStds[j] > 1e-8 {
				lr.Weights[j] = lr.Weights[j] * lr.yStd / lr.xStds[j]
			}
		}
		biasTerm := lr.yMean
		for j := 0; j < numFeatures; j++ {
			biasTerm -= lr.Weights[j] * lr.xMeans[j]
		}
		lr.Bias = lr.Bias*lr.yStd + biasTerm
	}
	return nil
}

func (lr *LinearRegression) calculateLossMultivariate(xs [][]float64, ys []float64) float64 {
	var totalLoss float64
	for i := range xs {
		pred := lr.Bias
		for j := range lr.Weights {
			pred += lr.Weights[j] * xs[i][j]
		}
		err := pred - ys[i]
		totalLoss += err * err
	}
	return totalLoss / float64(len(xs))
}

func (lr *LinearRegression) GetTrainingMetrics() (loss []float64, converged bool) {
	return lr.TrainingLoss, lr.Converged
}

func (lr *LinearRegression) Evaluate(xs [][]float64, ys []float64) (r2, mse, rmse float64) {
	if len(xs) != len(ys) || len(xs) == 0 {
		return 0, 0, 0
	}
	predictions := lr.PredictBatch(xs)
	yMean := 0.0
	for _, y := range ys {
		yMean += y
	}
	yMean /= float64(len(ys))
	var ssRes, ssTot float64
	for i := range ys {
		ssRes += math.Pow(ys[i]-predictions[i], 2)
		ssTot += math.Pow(ys[i]-yMean, 2)
	}
	if ssTot > 0 {
		r2 = 1 - ssRes/ssTot
	}
	mse = ssRes / float64(len(ys))
	rmse = math.Sqrt(mse)
	return r2, mse, rmse
}

func (lr *LinearRegression) GetWeights() []float64 {
	weights := make([]float64, len(lr.Weights))
	copy(weights, lr.Weights)
	return weights
}

func (lr *LinearRegression) GetBias() float64 {
	return lr.Bias
}

func (lr *LinearRegression) ImportModelFromString(modelStr string) error {
	lines := strings.Split(modelStr, "\n")
	var (
		biasPattern    = regexp.MustCompile(`^Bias:\s*([0-9.\-eE]+)`)
		weightPattern  = regexp.MustCompile(`^W\d+:\s*([0-9.\-eE]+)`)
	)
	var (
		bias    float64
		weights []float64
		foundBias bool
	)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if m := biasPattern.FindStringSubmatch(line); m != nil {
			val, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				return fmt.Errorf("error parsing bias: %w", err)
			}
			bias = val
			foundBias = true
		} else if m := weightPattern.FindStringSubmatch(line); m != nil {
			val, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				return fmt.Errorf("error parsing weight: %w", err)
			}
			weights = append(weights, val)
		}
	}
	if !foundBias {
		return fmt.Errorf("bias not found in model string")
	}
	if len(weights) == 0 {
		return fmt.Errorf("no weights found in model string")
	}
	lr.Bias = bias
	lr.Weights = make([]float64, len(weights))
	copy(lr.Weights, weights)
	return nil
}

