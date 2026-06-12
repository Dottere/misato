package config

import (
	"encoding/json"
	"log"
	"os"
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
	ServerPort     *int   `json:"server_port"`
	FilesDir       string `json:"files_dir"`
	DebugMode      bool   `json:"debug_mode"`
	BindAddress    string `json:"bind_address"`
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	IdleTimeout    int    `json:"idle_timeout"`
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
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
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
