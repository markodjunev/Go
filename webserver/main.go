package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Car struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

var (
	cars  = []Car{}
	mutex = &sync.Mutex{}
	id    = 1
)

func main() {
	err := loadCarsFromJSON("cars.json")
	if err != nil {
		fmt.Println("Error loading cars from JSON:", err);
		return;
	}

	if len(cars) > 0 {
		lastID, _ := strconv.Atoi(cars[len(cars)-1].ID);
		id = lastID + 1;
	}

	mux := http.NewServeMux();

	mux.HandleFunc("/cars", handleCars)       // POST, GET all
	mux.HandleFunc("/cars/", handleCarByID)  // GET, PUT, DELETE by ID

	fmt.Println("Starting server at port 8080");

	err = http.ListenAndServe(":8080", mux);
	if err != nil {
		fmt.Println("Error starting server: ", err);
	}
}

func loadCarsFromJSON(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &cars)
	if err != nil {
		return err
	}

	return nil
}

func handleCars(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createCar(w, r);
	case http.MethodGet:
		getAllCars(w);
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed);
	}
}

func handleCarByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/cars/"):];
	switch r.Method {
	case http.MethodGet:
		getCarByID(w, id);
	case http.MethodPut:
		updateCar(w, r, id);
	case http.MethodDelete:
		deleteCar(w, id);
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed);
	}
}

func createCar(w http.ResponseWriter, r *http.Request) {
	var car Car;
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	car.ID = strconv.Itoa(id);
	id++;
	
	cars = append(cars, car);
	w.WriteHeader(http.StatusCreated);
}

func getAllCars(w http.ResponseWriter) {
	mutex.Lock();
	defer mutex.Unlock();

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json");
	json.NewEncoder(w).Encode(cars);
}

func getCarByID(w http.ResponseWriter, id string) {
	mutex.Lock();
	defer mutex.Unlock();

	exists, car := exist(id)
	if !exists {
		http.Error(w, "Car not found", http.StatusNotFound);
		return;
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json");
	json.NewEncoder(w).Encode(car);
}

func updateCar(w http.ResponseWriter, r *http.Request, id string) {
	mutex.Lock();
	defer mutex.Unlock();

	var updatedCar Car
	if err := json.NewDecoder(r.Body).Decode(&updatedCar); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	exists, _ := exist(id)
	if !exists {
		http.Error(w, "Car not found", http.StatusNotFound)
		return
	}

	for i := range cars {
		if cars[i].ID == id {
			updatedCar.ID = id
			cars[i] = updatedCar
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json");
	json.NewEncoder(w).Encode(updatedCar);
}

func deleteCar(w http.ResponseWriter, id string) {
	mutex.Lock();
	defer mutex.Unlock();

	exist, _ := exist(id);

	if !exist {
		http.Error(w, "Car not found on deleting", http.StatusNotFound);
		return;
	}

	for i, car := range cars {
		if car.ID == id {
			cars = append(cars[:i], cars[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusOK);
}

func exist(id string) (bool, *Car) {
	for _, car := range cars {
		if car.ID == id {
			return true, &car
		}
	}

	return false, nil;
}