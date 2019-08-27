package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type weatherProvider interface {
	temperature(city string) (float64, error)
}

type openWeatherMap struct {}

func main() {
	http.HandleFunc("/hello", hello)

	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		city := strings.SplitN(r.URL.Path, "/", 3)[2]
		data, err := query(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(data)
	})

	http.ListenAndServe(":8080", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func query(city string) (weatherData, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=c286058d5f9666856bdbf3688600114f&q=" + city)
	if err != nil {
		return weatherData{}, err
	}
	
	defer resp.Body.Close()

	var d weatherData
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}
	
	return d, nil
}

func (w openWeatherMap) temperature(city string) (float64, error){
	
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=c286058d5f9666856bdbf3688600114f&q=" + city)
	if err != nil {
		return 0, err
	}
	
	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}
	
	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}

type weatherUnderground struct {
	apiKey string
}

func (w weatherUnderground) temperature(city string) (float64, error){
	
	resp, err := http.Get("http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/")
	if err != nil {
		return 0, err
	}
	
	defer resp.Body.Close()

	var d struct {
		Observation struct {
			Celsius float64 `json:"temp_c"`
		} `json:"current_observation"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}
	
	kelvin := d.Observation.Celsius + 273.15
	log.Printf("weatherUnderground: %s: %.2f", city, Kelvin)
	return Kelvin, nil
}

func temperature(city string, providers ...weatherProvider) (float64, error) {
	sum := 0.0
	for _, provider := range providers {
		k, err := provider.temperature(city)
		if err != nil {
			return 0, err
		}
		sum += k
	}
	return sum / float64(len(providers)), nil
}

type multiWeatherProvider []weatherProvider

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	sum := 0.0
	for _, provider := range w {
		k, err := provider.temperature(city)
		if err != nil {
			return 0, err
		}
		sum += k
	}
	return sum / float64(len(w)), nil
}
