package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"vpn-log-collector/internal/auth"
	"vpn-log-collector/internal/vpn"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Auth struct {
		APIToken    string `yaml:"api_token"`
		TokenHeader string `yaml:"token_header"`
	} `yaml:"auth"`
	Logging struct {
		Level    string `yaml:"level"`
		FilePath string `yaml:"file_path"`
	} `yaml:"logging"`
	VPN struct {
		LogdbaPath   string `yaml:"logdba_path"`
		LogDBPath    string `yaml:"logdb_path"`
		DefaultFlags string `yaml:"default_flags"`
		Timeout      int    `yaml:"timeout"`
	} `yaml:"vpn"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func loadConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Auth.TokenHeader == "" {
		config.Auth.TokenHeader = "X-API-Token"
	}

	return config, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Status:  "success",
		Message: "Service is healthy",
		Data:    map[string]string{"version": "1.0.0"},
	})
}

func vpnLogsHandler(client *vpn.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dateFrom := r.URL.Query().Get("from")
		dateTo := r.URL.Query().Get("to")

		if dateFrom == "" || dateTo == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "error",
				Message: "Missing 'from' or 'to' query parameters",
			})
			return
		}

		logs, err := client.GetLogs(dateFrom, dateTo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "error",
				Message: fmt.Sprintf("Failed to get logs: %v", err),
			})
			return
		}

		json.NewEncoder(w).Encode(Response{
			Data: logs,
		})
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if config.Logging.FilePath != "" {
		logFile, err := os.OpenFile(config.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Warning: could not open log file: %v. Using stdout.", err)
		} else {
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	} else {
		log.SetOutput(os.Stdout)
	}

	vpnClient := vpn.NewClient(
		config.VPN.LogdbaPath,
		config.VPN.LogDBPath,
		config.VPN.Timeout,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/vpn/logs", vpnLogsHandler(vpnClient))

	authenticatedMux := auth.TokenAuthMiddleware(config.Auth.APIToken, config.Auth.TokenHeader)(mux)

	server := &http.Server{
		Addr:    config.Server.Host + ":" + config.Server.Port,
		Handler: authenticatedMux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s:%s", config.Server.Host, config.Server.Port)
		log.Printf("API Token: %s...", config.Auth.APIToken[:min(8, len(config.Auth.APIToken))])

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
	log.Println("Server stopped")
}
