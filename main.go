package main

import (
	"backend/lr"
	"backend/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
)

var modelData string
var modelMutex sync.RWMutex

func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func predictHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type PredictionResult struct {
		Prediction float64 `json:"prediction"`
		Error      string  `json:"error,omitempty"`
	}

	resultChan := make(chan PredictionResult, 1)

	go func() {
		defer close(resultChan)

		features, err := utils.ParseStudentDataToFeatures(body)
		if err != nil {
			resultChan <- PredictionResult{Error: "Invalid JSON or feature extraction failed: " + err.Error()}
			return
		}

		model := lr.New(utils.GetExpectedFeatureCount())

		modelMutex.RLock()
		currentModelData := modelData
		modelMutex.RUnlock()

		err = model.ImportModelFromString(currentModelData)
		if err != nil {
			resultChan <- PredictionResult{Error: "Failed to load model: " + err.Error()}
			return
		}

		expectedWeights := len(model.GetWeights())
		fmt.Println(features)
		if len(features) != expectedWeights {
			resultChan <- PredictionResult{Error: fmt.Sprintf("Feature length mismatch: expected %d, got %d. Model was trained with %d features but current feature extraction produces %d features. Please retrain the model.", expectedWeights, len(features), expectedWeights, len(features))}
			return
		}

		prediction := model.Predict(features)
		roundedPrediction := math.Round(prediction*100) / 100
		resultChan <- PredictionResult{Prediction: roundedPrediction}
	}()

	result := <-resultChan

	w.Header().Set("Content-Type", "application/json")
	if result.Error != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": result.Error})
	} else {
		json.NewEncoder(w).Encode(map[string]float64{"prediction": result.Prediction})
	}
}

func loadModelData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	modelMutex.Lock()
	modelData = string(data)
	modelMutex.Unlock()
	
	return nil
}

func main() {
	err := loadModelData("values.txt")
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}
	fmt.Println("Model loaded successfully.")

	http.HandleFunc("/predict", predictHandler)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Printf("Expected feature count: %d\n", utils.GetExpectedFeatureCount())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
