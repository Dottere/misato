package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Config az alkalmazás működéséhez szükséges paramétereket tartalmazza.
// Az értékeket elsődlegesen a JSON konfigurációs fájlból olvassa be,
// de bizonyos mezőket (pl. port) a parancssori argumentumok felülírhatnak.
type Config struct {
	// ConfigFilePath a beolvasott fájl pontos helye. Parancssorból (-c) állítható.
	ConfigFilePath string
	// VerboseMode határozza meg, hogy részletes (debug-jellegű) logok jelenjenek-e meg induláskor.
	VerboseMode bool
	// ServerPort a HTTP szerver hallgatási portja. Parancssorból (-p) felülírható.
	ServerPort int `json:"server_port"`
	// FilesDir a mappa elérési útja, ahol a feldolgozandó manga (.cbz) fájlok találhatók.
	FilesDir string `json:"files_dir"`
	// DebugMode bekapcsolása esetén a HTML sablonok nem lesznek gyorsítótárazva.
	DebugMode bool `json:"debug_mode"`
	// AllowGuestAccess engedélyezi a könyvtár olvasását bejelentkezés (fiók) nélkül is.
	AllowGuestAccess bool `json:"allow_guest_access"`
	// BindAddress az IP cím (pl. "0.0.0.0" vagy "127.0.0.1"), amihez a szerver hozzáköti magát.
	BindAddress string `json:"bind_address"`
	// ReadTimeout a maximális idő, amit a szerver egy kérés beolvasására vár.
	ReadTimeout ConfigDuration `json:"read_timeout"`
	// WriteTimeout a maximális idő, amit a szerver a válasz elküldésére szán.
	WriteTimeout ConfigDuration `json:"write_timeout"`
	// IdleTimeout a maximális inaktív idő a Keep-Alive kapcsolatoknál.
	IdleTimeout ConfigDuration `json:"idle_timeout"`
	// DBPath az SQLite adatbázisfájl elérési útja (ha nem létezik, a szerver létrehozza).
	DBPath string `json:"database_path"`
}

// ConfigDuration a beépített time.Duration csomag-lokális burkolója (wrapper).
// Lehetővé teszi, hogy egyedi JSON feldolgozási logikát írjunk az időtartamokhoz.
type ConfigDuration time.Duration

// UnmarshalJSON feldolgozza a JSON fájlban megadott olvasható időformátumot (pl. "10s", "2m").
func (d *ConfigDuration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = ConfigDuration(parsed)
	return nil
}

// MarshalJSON visszaalakítja az időtartamot ember által olvasható formátumba a JSON mentéshez.
func (d ConfigDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// loadConfig beolvassa és feldolgozza a megadott JSON konfigurációs fájlt.
// Ha a fájl nem létezik, létrehoz egyet az alapértelmezett értékekkel,
// és azokkal tér vissza, elkerülve a program leállását.
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

// SetupConfig összefésüli a lemezről beolvasott konfigurációt a parancssori paraméterekkel.
// A parancssorban megadott értékek (pl. port) prioritást élveznek a fájlban lévőkkel szemben.
func SetupConfig(configFilePath string, cliPort int, verboseMode bool) (Config, error) {
	cfg, err := loadConfig(configFilePath)
	if err != nil {
		return Config{}, err
	}

	if cliPort != 0 {
		cfg.ServerPort = cliPort
	} else if cfg.ServerPort == 0 {
		cfg.ServerPort = 8080 // fallback
	}

	cfg.ConfigFilePath = configFilePath
	cfg.VerboseMode = verboseMode

	return *cfg, nil
}

// createDefaultConfigFile generál egy alapértelmezett konfigurációs fájlt a lemezen,
// ha az induláskor nem található meglévő példány.
func createDefaultConfigFile() {
	defaultConfig := Config{
		ConfigFilePath: "config.json",
		ServerPort:     8080,
		FilesDir:       "mangas",
		DebugMode:      false,
		BindAddress:    "0.0.0.0",
		ReadTimeout:    ConfigDuration(10 * time.Second),
		WriteTimeout:   ConfigDuration(10 * time.Second),
		IdleTimeout:    ConfigDuration(120 * time.Second),
		DBPath:         "misato.db",
	}

	json, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		log.Fatalf("Error when creating default config: %v", err)
	}

	os.WriteFile("config.json", json, 0644)

}
