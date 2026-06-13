package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

/*
A config releváns paramétereit tartalmazó struktúra, csak ezeket olvassa ki a JSON fájlból.

# Paraméterek
  - ConfigFilePath: A config fájl elérési útja, akkor állítódik ha a szerver indításakor -c vagy --config flaggel beállítják
  - ServerPort: A port amin a szerver fut, kötelező megadni, vagy a -p/--port flagekkel vagy pedig a configban
  - FilesDir: A mappa ahol a mangák tárolva vannak, a webszerver innen olvassa ki őket
  - DebugMode: Beállítja, hogy a szerver debug üzemmódban fut-e !MÉG NEM CSINÁL SEMMIT, TO BE IMPLEMENTED!
  - BindAddress: Az IP cím amihez a szervert hozzákötjük
  - ReadTimeout: Ha egy olvasási kérelem túl sokáig tart akkor ennyi másodperc után kidob a szerver
  - WriteTimeout: Ha egy írási kérelem túl sokáig tart, akkor ennyi másodperc után kidob a szerver
  - IdleTimeout: Ha a felhasználó nem csinál semmit, akkor ennyi másodperc után kidob a szerver
*/
type Config struct {
	ConfigFilePath string
	ServerPort     *int           `json:"server_port"`
	FilesDir       string         `json:"files_dir"`
	DebugMode      bool           `json:"debug_mode"`
	BindAddress    string         `json:"bind_address"`
	ReadTimeout    ConfigDuration `json:"read_timeout"`
	WriteTimeout   ConfigDuration `json:"write_timeout"`
	IdleTimeout    ConfigDuration `json:"idle_timeout"`
}

// Mivel a beépített típusokat nem lehet
// módosítani ezért csinálunk egy sajátot ami
// kvázi ugyanaz mint a time.Duration csak
// így hozzá lehet akasztani egy új metódust
type ConfigDuration time.Duration

// Mivel a time.Duration típus nem rendelkezik UnmarshalJSON implementációval
// ezért itt definiálunk egy sajátot, így konfigurációfájlban megadhatóvá
// válik ez a része is a kódnak.
func (d *ConfigDuration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = ConfigDuration(parsed)
	return nil
}

/*
Populál egy Config struktúrát és visszaadja azt.
Ehhez beolvassa a JSON fájlt és a beépített JSON parser
segítségével értelmezi azt.
*/
func loadConfig(configFilePath string) (*Config, error) {

	if configFilePath == "" {
		configFilePath = "config.json"
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &cfg, nil
}

func SetupConfig(configFilePath string, cliPort int) Config {
	cfg, err := loadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	finalPort := cliPort

	if finalPort == 0 {
		if cfg.ServerPort == nil {
			log.Fatal("Error: Configuration is missing port field and no CLI port provided.")
		}
		finalPort = *cfg.ServerPort
	} else {
		if cfg.ServerPort == nil {
			cfg.ServerPort = new(int)
		}
		*cfg.ServerPort = finalPort
	}

	cfg.ConfigFilePath = configFilePath

	return *cfg
}
