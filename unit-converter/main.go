package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	router := gin.Default()

	// Serve static files (like CSS)
	router.Static("/static", "./static")

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"result": "",
		})
	})

	// Handle form submission
	router.POST("/convert", func(c *gin.Context) {
		// Get form values
		parameter := c.PostForm("parameter")
		value := c.PostForm("value")
		fromUnit := c.PostForm("from_unit")
		toUnit := c.PostForm("to_unit")

		// Convert string to float
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Invalid number entered",
				"result": "",
			})
			return
		}

		// Perform conversion based on parameter type
		result := convertUnits(val, fromUnit, toUnit, parameter)
		
		// Check for conversion error
		if result == "Unsupported unit" || result == "Unsupported parameter" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": result,
				"result": "",
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"result": result,
			"original_value": value,
			"from_unit": fromUnit,
			"to_unit": toUnit,
			"parameter": parameter,
		})
	})

	// Start server
	router.Run(":8080") // Visit http://localhost:8080
}

func convertUnits(value float64, fromUnit, toUnit, parameter string) string {
	parameter = strings.ToLower(parameter)
	// Conversion logic
	switch parameter {
	case "length":
		return convertLength(value, fromUnit, toUnit)
	case "temperature":
		return convertTemperature(value, fromUnit, toUnit)
	case "weight":
		return convertWeight(value, fromUnit, toUnit)
	default:
		return "Unsupported parameter"
	}
}

// Map for length conversions
var lengthConversions = map[string]float64{
	"meters":      1.0,
	"feet":        0.3048,
	"inches":      0.0254,
	"kilometers":  1000.0,
	"centimeters": 0.01,
	"millimeters": 0.001,
}

//Convert lenght units
func convertLength(value float64, fromUnit string, toUnit string) string {
	fromValue, okFrom := lengthConversions[fromUnit]
	toValue, okTo := lengthConversions[toUnit]
	if !okFrom || !okTo {
		return "Unsupported unit"
	}

	// Convert to meters first
	valueInMeters := value * fromValue

	// Convert from meters to the target unit
	result := valueInMeters / toValue
	
	return strconv.FormatFloat(result, 'f', 2, 64) + " " + toUnit
}

// Map for temperature conversions
var temperatureConversions = map[string]float64{
	"celsius":    0.0,
	"fahrenheit": 5.0 / 9.0,
	"kelvin":     273.15,
}
// Convert temperature units
func convertTemperature(value float64, fromUnit string, toUnit string) string {
	fromValue, okFrom := temperatureConversions[fromUnit]
	toValue, okTo := temperatureConversions[toUnit]
	if !okFrom || !okTo {
		return "Unsupported unit"
	}
	
	// Convert to Celsius first
	if fromUnit == "fahrenheit" {
		value = (value - 32) * fromValue
	} else if fromUnit == "kelvin" {
		value -= fromValue
	}
	// Convert from Celsius to the target unit
	if toUnit == "fahrenheit" {
		value = value/toValue + 32
	} else if toUnit == "kelvin" {
		value += toValue
	} else {
		value *= toValue
	}

	return strconv.FormatFloat(value, 'f', 2, 64) + " " + toUnit
}

// Map for weight conversions
var weightConversions = map[string]float64{
	"kilograms":  1.0,
	"grams":      0.001,
	"pounds":     0.453592,
	"ounces":     0.0283495,
	"tons":      907.185,
	"stones":     6.35029,
}

// Convert weight units
func convertWeight(value float64, fromUnit string, toUnit string) string {
	fromValue, okFrom := weightConversions[fromUnit]
	toValue, okTo := weightConversions[toUnit]
	if !okFrom || !okTo {
		return "Unsupported unit"
	}

	// Convert to kilograms first
	valueInKilograms := value * fromValue

	// Convert from kilograms to the target unit
	result := valueInKilograms / toValue
	
	return strconv.FormatFloat(result, 'f', 2, 64) + " " + toUnit
}