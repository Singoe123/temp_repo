package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseDateToFloat(dateStr string) (float64, error) {
	dateStr = strings.TrimSpace(dateStr)
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("formato de fecha inválido: %v", err)
	}

	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	daysSinceEpoch := date.Sub(epoch).Hours() / 24

	return daysSinceEpoch, nil
}

type StudentData struct {
	CicloAcademico                        string `json:"CICLO_ACADEMICO"`
	FechaMatricula                        string `json:"FECHA_MATRICULA"`
	PeriodoAcademicoAnterior              string `json:"PERIODO_ACADEMICO_ANTERIOR"`
	CreditosAcumuladosAprobadosAnterior   string `json:"CREDITOS_ACUMULADOS_APROBADOS_AL_PERIODO_ANTERIOR"`
	CreditosMatriculadosAnterior          string `json:"CREDITOS_MATRICULADOS_DEL_PERIODO_ANTERIOR"`
	CreditosAprobadosAnterior             string `json:"CREDITOS_APROBADOS_DEL_PERIODO_ANTERIOR"`
	Genero                                string `json:"GENERO"`
	Discapacidad                          string `json:"DISCAPACIDAD"`
	Programa                              string `json:"PROGRAMA"`
	Facultad                              string `json:"FACULTAD"`
	Edad                                  string `json:"Edad"`
}

func ParseStudentDataToFeatures(jsonData []byte) ([]float64, error) {
	var data StudentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	features := make([]float64, 0, 60)

	ciclo, _ := strconv.ParseFloat(data.CicloAcademico, 64)
	features = append(features, ciclo)

	fechaMatricula, err := parseDateToFloat(data.FechaMatricula)
	if err != nil {
		fechaMatricula = 0.0
	}
	features = append(features, fechaMatricula)

	periodoStr := strings.Map(func(r rune) rune {
		if (r == '-') {
			return -1
		}
		return r
	}, data.PeriodoAcademicoAnterior)
	periodo, _ := strconv.ParseFloat(periodoStr, 64)
	features = append(features, periodo)

	credAcum, _ := strconv.ParseFloat(data.CreditosAcumuladosAprobadosAnterior, 64)
	credMat, _ := strconv.ParseFloat(data.CreditosMatriculadosAnterior, 64)
	credAprob, _ := strconv.ParseFloat(data.CreditosAprobadosAnterior, 64)
	features = append(features, credAcum, credMat, credAprob)

	if data.Genero == "Masculino" {
		features = append(features, 1.0)
	} else {
		features = append(features, 0.0)
	}

	if data.Discapacidad == "Si" {
		features = append(features, 1.0)
	} else {
		features = append(features, 0.0)
	}

	edad, _ := strconv.ParseFloat(data.Edad, 64)
	features = append(features, edad)

	programas := []string{
		"AGRONOMIA", "ARQUITECTURA Y URBANISMO", "BIOLOGIA", "CIENCIAS ADMINISTRATIVAS",
		"CIENCIAS CONTABLES Y FINANCIERAS", "CIENCIAS DE LA COMUNICACION", "DERECHO Y CIENCIAS POLITICAS",
		"ECONOMIA", "EDUCACION INICIAL", "EDUCACION PRIMARIA", "ELECTRONICA Y TELECOMUNICACIONES",
		"ENFERMERIA", "ESPECIALIDAD EN ADMINISTRACIÓN", "ESTADISTICA", "ESTOMATOLOGIA", "FISICA",
		"HISTORIA Y GEOGRAFIA", "INGENIERIA AGRICOLA", "INGENIERIA AGROINDUSTRIAL E INDUSTRIAS ALIMENTARIAS",
		"INGENIERIA AMBIENTAL Y SEGURIDAD INDUSTRIAL", "INGENIERIA CIVIL", "INGENIERIA DE MINAS",
		"INGENIERIA DE PETROLEO", "INGENIERIA GEOLOGICA", "INGENIERIA INDUSTRIAL", "INGENIERIA INFORMATICA",
		"INGENIERIA MECATRONICA", "INGENIERIA PESQUERA", "INGENIERIA QUIMICA", "LENGUA Y LITERATURA",
		"MATEMATICA", "MEDICINA HUMANA", "MEDICINA VETERINARIA", "OBSTETRICIA", "PSICOLOGIA", "ZOOTECNIA",
	}
	for i := range programas {
		if data.Programa == programas[i] {
			features = append(features, 1.0)
		} else {
			features = append(features, 0.0)
		}
	}

	facultades := []string{
		"AGRONOMIA", "ARQUITECTURA Y URBANISMO", "CIENCIAS", "CIENCIAS ADMINISTRATIVAS",
		"CIENCIAS CONTABLES Y FINANCIERAS", "CIENCIAS DE LA SALUD", "CIENCIAS SOCIALES Y EDUCACION",
		"DERECHO Y CIENCIAS POLITICAS", "ECONOMIA", "INGENIERIA CIVIL", "INGENIERIA DE MINAS",
		"INGENIERIA INDUSTRIAL", "INGENIERIA PESQUERA ", "PROGRAMA DE COMPLEMENTACIÓN ACADÉMICO PROFESIONAL EN ADMINISTRACIÓN- CONVENIO IPAE",
		"ZOOTECNIA",
	}
	for i := range facultades {
		if data.Facultad == facultades[i] {
			features = append(features, 1.0)
		} else {
			features = append(features, 0.0)
		}
	}

	return features, nil
}

func GetFeatureNames() []string {
	names := []string{
		"CICLO_ACADEMICO",
		"FECHA_MATRICULA",
		"PERIODO_ACADEMICO_ANTERIOR",
		"CREDITOS_ACUMULADOS_APROBADOS_AL_PERIODO_ANTERIOR",
		"CREDITOS_MATRICULADOS_DEL_PERIODO_ANTERIOR",
		"CREDITOS_APROBADOS_DEL_PERIODO_ANTERIOR",
		"GENERO",
		"DISCAPACIDAD",
		"Edad",
	}

	programas := []string{
		"Programa_AGRONOMIA", "Programa_ARQUITECTURA Y URBANISMO", "Programa_BIOLOGIA",
		"Programa_CIENCIAS ADMINISTRATIVAS", "Programa_CIENCIAS CONTABLES Y FINANCIERAS",
		"Programa_CIENCIAS DE LA COMUNICACION", "Programa_DERECHO Y CIENCIAS POLITICAS",
		"Programa_ECONOMIA", "Programa_EDUCACION INICIAL", "Programa_EDUCACION PRIMARIA",
		"Programa_ELECTRONICA Y TELECOMUNICACIONES", "Programa_ENFERMERIA",
		"Programa_ESPECIALIDAD EN ADMINISTRACIÓN", "Programa_ESTADISTICA", "Programa_ESTOMATOLOGIA",
		"Programa_FISICA", "Programa_HISTORIA Y GEOGRAFIA", "Programa_INGENIERIA AGRICOLA",
		"Programa_INGENIERIA AGROINDUSTRIAL E INDUSTRIAS ALIMENTARIAS",
		"Programa_INGENIERIA AMBIENTAL Y SEGURIDAD INDUSTRIAL", "Programa_INGENIERIA CIVIL",
		"Programa_INGENIERIA DE MINAS", "Programa_INGENIERIA DE PETROLEO", "Programa_INGENIERIA GEOLOGICA",
		"Programa_INGENIERIA INDUSTRIAL", "Programa_INGENIERIA INFORMATICA", "Programa_INGENIERIA MECATRONICA",
		"Programa_INGENIERIA PESQUERA", "Programa_INGENIERIA QUIMICA", "Programa_LENGUA Y LITERATURA",
		"Programa_MATEMATICA", "Programa_MEDICINA HUMANA", "Programa_MEDICINA VETERINARIA",
		"Programa_OBSTETRICIA", "Programa_PSICOLOGIA", "Programa_ZOOTECNIA",
	}
	names = append(names, programas...)

	facultades := []string{
		"Facultad_AGRONOMIA", "Facultad_ARQUITECTURA Y URBANISMO", "Facultad_CIENCIAS",
		"Facultad_CIENCIAS ADMINISTRATIVAS", "Facultad_CIENCIAS CONTABLES Y FINANCIERAS",
		"Facultad_CIENCIAS DE LA SALUD", "Facultad_CIENCIAS SOCIALES Y EDUCACION",
		"Facultad_DERECHO Y CIENCIAS POLITICAS", "Facultad_ECONOMIA", "Facultad_INGENIERIA CIVIL",
		"Facultad_INGENIERIA DE MINAS", "Facultad_INGENIERIA INDUSTRIAL", "Facultad_INGENIERIA PESQUERA ",
		"Facultad_PROGRAMA DE COMPLEMENTACIÓN ACADÉMICO PROFESIONAL EN ADMINISTRACIÓN- CONVENIO IPAE",
		"Facultad_ZOOTECNIA",
	}
	names = append(names, facultades...)

	return names
}

func GetExpectedFeatureCount() int {
	return len(GetFeatureNames())
}

func TestStudentDataTransformation() {
	jsonData := `{
		"CICLO_ACADEMICO": "6.0",
		"FECHA_MATRICULA": "2025-03-26",
		"PERIODO_ACADEMICO_ANTERIOR": "20242.0",
		"CREDITOS_ACUMULADOS_APROBADOS_AL_PERIODO_ANTERIOR": "123.0",
		"CREDITOS_MATRICULADOS_DEL_PERIODO_ANTERIOR": "18.0",
		"CREDITOS_APROBADOS_DEL_PERIODO_ANTERIOR": "18.0",
		"PROGRAMA": "AGRONOMIA",
		"FACULTAD": "AGRONOMIA",
		"GENERO": "Masculino",
		"DISCAPACIDAD": "No",
		"Edad": "21.232032854209447"
	}`

	features, err := ParseStudentDataToFeatures([]byte(jsonData))
	if err != nil {
		panic(err)
	}

	featureNames := GetFeatureNames()
	
	fmt.Println("Feature transformation example (matching CSV structure):")
	fmt.Printf("Total features: %d\n", len(features))
	fmt.Printf("Expected features: %d\n", len(featureNames))
	fmt.Println("First 9 features:")
	
	for i := 0; i < Min(9, len(features)); i++ {
		if i < len(featureNames) {
			fmt.Printf("  %s: %.2f\n", featureNames[i], features[i])
		}
	}

	fmt.Println("\nPrograma dummy variables (all 0 in this example):")
	for i := 9; i < Min(19, len(features)); i++ {
		if i < len(featureNames) {
			fmt.Printf("  %s: %.0f\n", featureNames[i], features[i])
		}
	}

	fmt.Println("\nFacultad dummy variables (all 0 in this example):")
	startFac := 9 + 36
	for i := startFac; i < Min(startFac+5, len(features)); i++ {
		if i < len(featureNames) {
			fmt.Printf("  %s: %.0f\n", featureNames[i], features[i])
		}
	}
	
	fmt.Println("\nTesting different gender and disability values:")
	
	jsonData2 := `{
		"CICLO_ACADEMICO": "4.0",
		"FECHA_MATRICULA": "2025-03-25",
		"PERIODO_ACADEMICO_ANTERIOR": "20241.0",
		"CREDITOS_ACUMULADOS_APROBADOS_AL_PERIODO_ANTERIOR": "80.0",
		"CREDITOS_MATRICULADOS_DEL_PERIODO_ANTERIOR": "20.0",
		"CREDITOS_APROBADOS_DEL_PERIODO_ANTERIOR": "18.0",
		"GENERO": "Femenino",
		"DISCAPACIDAD": "Si",
		"Edad": "19.5"
	}`

	features2, err := ParseStudentDataToFeatures([]byte(jsonData2))
	if err != nil {
		panic(err)
	}

	fmt.Printf("GENERO (Femenino): %.0f\n", features2[6])
	fmt.Printf("DISCAPACIDAD (Si): %.0f\n", features2[7])
}