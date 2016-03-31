package main // import "github.com/Jimdo/vault-unseal"

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Luzifer/rconfig"
)

var config = struct {
	OneShot       bool   `flag:"oneshot,1" default:"false" description:"Only try once and exit after"`
	SealTokensRaw string `flag:"tokens" default:"" description:"Tokens to try for unsealing the vault instance"`
	SealTokens    []string
	VaultInstance string `flag:"instance" env:"VAULT_ADDR" default:"http://127.0.0.1:8200" description:"Vault instance to unlock"`
	Sleep         int    `flag:"sleep" default:"30" description:"How long to wait between sealed-state checks"`
}{}

func init() {
	if err := rconfig.Parse(&config); err != nil {
		log.Printf("Unable to parse CLI parameters: %s\n", err)
		os.Exit(1)
	}

	if len(config.SealTokensRaw) == 0 {
		log.Println("You must provide at least one token.")
		os.Exit(1)
	}

	config.SealTokens = strings.Split(config.SealTokensRaw, ",")
}

func main() {
	for {
		s := sealStatus{}
		r, err := http.Get(config.VaultInstance + "/v1/sys/seal-status")
		if err != nil {
			log.Printf("An error ocurred while reading seal-status: %s\n", err)
			os.Exit(1)
		}
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			log.Printf("Unable to decode seal-status: %s\n", err)
			os.Exit(1)
		}

		if s.Sealed {
			for _, token := range config.SealTokens {
				log.Printf("Vault instance is sealed (missing %d tokens), trying to unlock...\n", s.T-s.Progress)
				body := bytes.NewBuffer([]byte{})
				json.NewEncoder(body).Encode(map[string]interface{}{
					"key": token,
				})
				r, _ := http.NewRequest("PUT", config.VaultInstance+"/v1/sys/unseal", body)
				resp, err := http.DefaultClient.Do(r)
				if err != nil {
					log.Printf("An error ocurred while doing unseal: %s\n", err)
					os.Exit(1)
				}
				defer resp.Body.Close()

				if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
					log.Printf("Unable to decode seal-status: %s\n", err)
					os.Exit(1)
				}

				if !s.Sealed {
					log.Printf("Unseal successfully finished.\n")
					break
				}
			}

			if s.Sealed {
				log.Printf("Vault instance is still sealed (missing %d tokens), I don't have any more tokens.\n", s.T-s.Progress)
			}
		}

		if config.OneShot {
			break
		} else {
			<-time.After(time.Duration(config.Sleep) * time.Second)
		}
	}
}
