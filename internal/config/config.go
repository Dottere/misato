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
  - VerboseMode: Meghatározza, hogy a szerver elindulásakor íródjanak-e ki üzenetek (jelenleg cli beviteli alias a debugra)
  - ServerPort: A port amin a szerver fut, kötelező megadni, vagy a -p/--port flagekkel vagy pedig a configban
  - FilesDir: A mappa ahol a mangák tárolva vannak, a webszerver innen olvassa ki őket
  - DebugMode: Beállítja, hogy a szerver debug üzemmódban fut-e
  - BindAddress: Az IP cím amihez a szervert hozzákötjük
  - ReadTimeout: Ha egy olvasási kérelem túl sokáig tart akkor ennyi másodperc után kidob a szerver
  - WriteTimeout: Ha egy írási kérelem túl sokáig tart, akkor ennyi másodperc után kidob a szerver
  - IdleTimeout: Ha a felhasználó nem csinál semmit, akkor ennyi másodperc után kidob a szerver
  - DBPath: Az adatbázisfájl elérési útja és neve, ha nem létezik itt hozza létre
*/
type Config struct {
	ConfigFilePath   string
	VerboseMode      bool
	ServerPort       int            `json:"server_port"`
	FilesDir         string         `json:"files_dir"`
	DebugMode        bool           `json:"debug_mode"`
	AllowGuestAccess bool           `json:"allow_guest_access"`
	BindAddress      string         `json:"bind_address"`
	ReadTimeout      ConfigDuration `json:"read_timeout"`
	WriteTimeout     ConfigDuration `json:"write_timeout"`
	IdleTimeout      ConfigDuration `json:"idle_timeout"`
	DBPath           string         `json:"database_path"`
}

// ConfigDuration a time.Duration egy csomag-lokális burkolója.
// Mivel idegen típusokhoz nem rögzíthetünk új metódusokat, szükség van egy
// saját típusra, hogy egyedi JSON beolvasási logikát írhassunk hozzá.
type ConfigDuration time.Duration

// UnmarshalJSON implementálja a json.Unmarshaler interfészt.
// Mivel a time.Duration típus alapból nem tudja string formátumból (pl. "10s")
// feldolgozni a JSON értékeket, ez a metódus teszi lehetővé, hogy konfigurációs
// fájlokban is emberi számára olvasható formában adhassuk meg az időtartamokat.
func (d *ConfigDuration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = ConfigDuration(parsed)
	return nil
}

func (d ConfigDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal((time.Duration(d) * time.Second).String())
}

// loadConfig beolvassa a megadott konfigurációs fájlt, és visszaadja a feldolgozott Config-ot.
// Ha a configFilePath üres, alapértelmezetten a "config.json"-t keresi.
// Ha a fájl nem létezik, automatikusan létrehoz egy alapértelmezett konfigurációs fájlt,
// és az abban lévő értékekkel tér vissza.
func loadConfig(configFilePath string) (*Config, error) {

	if configFilePath == "" {
		configFilePath = "config.json"
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		createDefaultConfigFile()

		data, err = os.ReadFile(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config after creation: %w", err)
		}
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &cfg, nil
}

// SetupConfig inicializálja az alkalmazás konfigurációját a fájl és a parancssori (CLI)
// paraméterek összefésülésével. A parancssorban megadott port mindig felülírja a fájlból
// beolvasott értéket.
//
// Figyelem: A függvény azonnal leállítja a program futását (log.Fatal), ha a fájl
// beolvasása sikertelen, vagy ha sem a parancssorban, sem a fájlban nem talált érvényes portot.
func SetupConfig(configFilePath string, cliPort int, verboseMode bool) (Config, error) {
	cfg, err := loadConfig(configFilePath)
	if err != nil {
		return Config{}, err
	}

	finalPort := cliPort

	if finalPort == 0 && cfg.ServerPort == 0 {
		cfg.ServerPort = 8080 // default is 8080
	}

	cfg.ConfigFilePath = configFilePath
	cfg.VerboseMode = verboseMode

	return *cfg, nil
}

func createDefaultConfigFile() {
	defaultConfig := Config{
		ConfigFilePath: "config.json",
		ServerPort:     8080,
		FilesDir:       "mangas",
		DebugMode:      false,
		BindAddress:    "0.0.0.0",
		ReadTimeout:    10,
		WriteTimeout:   10,
		IdleTimeout:    120,
		DBPath:         "misato.db",
	}

	json, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		log.Fatalf("Error when creating default config: %v", err)
	}

	os.WriteFile("config.json", json, 0644)

}
