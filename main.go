package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	// default values for configuration
	defaultEndpoint       = "es6:9200"
	defaultRepo           = "default"
	defaultCleanAfterDays = 7
	defaultDryRun         = false
)

// Config holds variable configuration which can be overriden using env var
type Config struct {
	endpoint, repo               string
	cleanAfterDays, keepMinSnaps int
	dryRun                       bool
}

// A Snapshot represent an elasticsearch snapshot as returned by the API
type Snapshot struct {
	Id       string `json:"id"`
	Start    string `json:"start_epoch"`
	End      string `json:"end_epoch"`
	Status   string `json:"status"`
	SuccessS string `json:"successful_shards"`
	FailedS  string `json:"failed_shards"`
}

func main() {
	config, err := getEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during environment initialization : %v\n", err)
	}

	listSnap, err := getSnaps(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during retrieval of snapshots : %v\n", err)
	}

	toClean, err := getToCleanSnaps(config, listSnap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during listing of snapshots to clean : %v\n", err)
	}
	err = cleanSnaps(config, listSnap, toClean)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during cleanup of snapshots : %v\n", err)
	}
}

func getEnv() (*Config, error) {
	config := &Config{
		endpoint:       defaultEndpoint,
		repo:           defaultRepo,
		cleanAfterDays: defaultCleanAfterDays,
		dryRun:         defaultDryRun,
	}
	if endpoint, set := os.LookupEnv("ES_ENDPOINT"); set {
		config.endpoint = endpoint
	}
	if repo, set := os.LookupEnv("ES_REPO"); set {
		config.repo = repo
	}
	if cleanAfterDays, set := os.LookupEnv("ES_CLEAN_AFTER_DAYS"); set {
		c, err := strconv.ParseInt(cleanAfterDays, 10, 16)
		if err != nil {
			return nil, err
		}
		config.cleanAfterDays = int(c)
	}
	if keepMinSnaps, set := os.LookupEnv("ES_KEEP_MIN_SNAPS"); set {
		k, err := strconv.ParseInt(keepMinSnaps, 10, 16)
		if err != nil {
			return nil, err
		}
		config.keepMinSnaps = int(k)
	}
	if _, set := os.LookupEnv("ES_DRY_RUN"); set {
		fmt.Println("Enabling dry-run mode")
		config.dryRun = true
	}
	return config, nil
}

func getSnaps(config *Config) ([]Snapshot, error) {
	resp, err := http.Get("http://" + config.endpoint + "/_cat/snapshots/" + config.repo + "?format=json")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("query to elasticsearch returned status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	listSnap := []Snapshot{}
	err = json.Unmarshal(body, &listSnap)
	if err != nil {
		return nil, err
	}
	return listSnap, nil
}

func getToCleanSnaps(config *Config, listSnap []Snapshot) ([]Snapshot, error) {
	toClean := []Snapshot{}
	for _, snap := range listSnap {
		old, err := isSnapOlderThan(config, snap)
		if err != nil {
			return nil, err
		}
		if old {
			toClean = append(toClean, snap)
		}
	}
	return toClean, nil
}

func isSnapOlderThan(config *Config, snap Snapshot) (bool, error) {
	snapEndEpoch, err := strconv.ParseInt(snap.End, 10, 64)
	if err != nil {
		return false, err
	}
	snapEnd := time.Unix(snapEndEpoch, 0)
	now := time.Now()
	snapDuration := now.Sub(snapEnd)
	cleanDuration := 24 * time.Hour * time.Duration(config.cleanAfterDays)
	if snapDuration > cleanDuration {
		return true, nil
	}
	return false, nil

}

func cleanSnaps(config *Config, listSnap []Snapshot, toClean []Snapshot) error {
	// If we don't have enough snapshots remaining after clean, abort
	remainingSnaps := len(listSnap) - len(toClean)
	fmt.Printf("%d snapshots to clean, remaining : %d\n", len(toClean), remainingSnaps)
	if remainingSnaps < config.keepMinSnaps {
		return fmt.Errorf("i'm about to remove all snapshots ! Exiting gracefully instead")
	}
	// Else we request a DELETE for every selected snapshot
	client := &http.Client{Timeout: 1 * time.Minute}
	for _, snap := range toClean {
		if config.dryRun {
			fmt.Printf("[DRY-RUN] Snap %v would have been deleted\n", snap.Id)
			continue
		}
		fmt.Printf("Deleting snap %v ...\n", snap.Id)
		req, err := http.NewRequest("DELETE", "http://"+config.endpoint+"/_snapshot/"+config.repo+"/"+snap.Id, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("error during deletion of snapshot %v with status : %v", snap.Id, resp.Status)
		}
	}
	return nil
}
