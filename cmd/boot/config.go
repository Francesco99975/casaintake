package boot

import (
	"fmt"
	"net"
	"os"

	"github.com/Francesco99975/casaintake/internal/enums"
)

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

type Config struct {
	Port  string
	Host  string
	GoEnv enums.Environment

	NTFY         string
	NTFYToken    string
	URL          string
	MetricSecret string
	Prometheus   string

	ResendAPIKey string
}

var Environment = &Config{}

func LoadEnvVariables() error {

	if !enums.IsEnvironmentValid(os.Getenv("GO_ENV")) {
		return fmt.Errorf("invalid environment variable: %s", os.Getenv("GO_ENV"))
	}

	Environment.Port = os.Getenv("PORT")
	Environment.Host = os.Getenv("HOST")
	Environment.GoEnv = enums.GetEnvironmentFromString(os.Getenv("GO_ENV"))

	Environment.NTFY = os.Getenv("NTFY")
	Environment.NTFYToken = os.Getenv("NTFY_TOKEN")
	Environment.MetricSecret = os.Getenv("METRIC_SECRET")
	Environment.Prometheus = os.Getenv("PROMETHEUS")
	if Environment.GoEnv == enums.Environments.DEVELOPMENT {
		localIP := getLocalIP()
		Environment.URL = fmt.Sprintf("http://%s:%s", localIP, Environment.Port)
	} else {
		Environment.URL = fmt.Sprintf("https://%s", Environment.Host)
	}

	Environment.ResendAPIKey = os.Getenv("RESEND_API_KEY")

	return nil
}
