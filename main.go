package main // import "github.com/Jimdo/vault-unseal"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/Luzifer/rconfig"
)

var config = struct {
	OneShot        bool     `flag:"oneshot,1" default:"false" description:"Only try once and exit after"`
	SealTokens     []string `flag:"tokens" default:"" description:"Tokens to try for unsealing the vault instance"`
	VaultInstances []string `flag:"instance" env:"VAULT_ADDR" default:"http://127.0.0.1:8200" description:"Vault instance to unlock"`
	Sleep          int      `flag:"sleep" default:"30" description:"How long to wait between sealed-state checks"`
}{}

func init() {
	if err := rconfig.Parse(&config); err != nil {
		log.Printf("Unable to parse CLI parameters: %s\n", err)
		os.Exit(1)
	}

	if len(config.SealTokens) == 1 && config.SealTokens[0] == "" {
		if len(rconfig.Args()) <= 1 {
			log.Println("You must provide at least one token.")
			os.Exit(1)
		}
		config.SealTokens = rconfig.Args()[1:]
	}
}

func main() {
	var wg sync.WaitGroup

	for {
		for i := range config.VaultInstances {
			wg.Add(1)
			go func(i int) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				defer wg.Done()
				defer cancel()

				if err := unsealInstance(ctx, config.VaultInstances[i]); err != nil {
					log.Printf("[ERR] %s", err)
				}
			}(i)
		}

		if config.OneShot {
			break
		} else {
			<-time.After(time.Duration(config.Sleep) * time.Second)
		}
	}

	wg.Wait()
}

func unsealInstance(ctx context.Context, instance string) error {
	s := sealStatus{}
	r, err := ctxhttp.Get(ctx, http.DefaultClient, instance+"/v1/sys/seal-status")
	if err != nil {
		return fmt.Errorf("[%s] An error ocurred while reading seal-status: %s", instance, err)
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		return fmt.Errorf("[%s] Unable to decode seal-status: %s", instance, err)
	}

	if s.Sealed {
		for _, token := range config.SealTokens {
			log.Printf("[%s] Vault instance is sealed (missing %d tokens), trying to unlock...", instance, s.T-s.Progress)
			body := bytes.NewBuffer([]byte{})
			json.NewEncoder(body).Encode(map[string]interface{}{
				"key": token,
			})
			r, _ := http.NewRequest("PUT", instance+"/v1/sys/unseal", body)
			resp, err := ctxhttp.Do(ctx, http.DefaultClient, r)
			if err != nil {
				return fmt.Errorf("[%s] An error ocurred while doing unseal: %s", instance, err)
			}
			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
				return fmt.Errorf("[%s] Unable to decode seal-status: %s", instance, err)
			}

			if !s.Sealed {
				log.Printf("[%s] Unseal successfully finished.", instance)
				break
			}
		}

		if s.Sealed {
			log.Printf("[%s] Vault instance is still sealed (missing %d tokens), I don't have any more tokens.", instance, s.T-s.Progress)
		}
	} else {
		log.Printf("[%s] Vault instance is already unsealed.", instance)
	}

	return nil
}
