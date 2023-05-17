package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"vault-controller/pkg/vault"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	EnvLogLevel = "LOG_LEVEL"
)

type VaultKubernetesRoleObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              vault.KubernetesRole `json:"spec"`
	Status            ControllerStatus     `json:"status"`
}

type VaultPolicyObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              vault.Policy     `json:"spec"`
	Status            ControllerStatus `json:"status"`
}

type ControllerStatus struct {
	Succeeded bool `json:"succeeded"`
}

type VaultKubernetesRoleRequest struct {
	Parent     VaultKubernetesRoleObject `json:"parent"`
	Finalizing bool                      `json:"finalizing"`
}

type VaultPolicyRequest struct {
	Parent     VaultPolicyObject `json:"parent"`
	Finalizing bool              `json:"finalizing"`
}

type SyncResponse struct {
	Status    ControllerStatus `json:"status"`
	Finalized bool             `json:"finalized"`
}

func init() {
	setLogLevel()
}

func main() {
	log.Info("Starting Vault Controller...")
	http.HandleFunc("/sync-kubernetes-role", syncKubernetesRoleHandler)
	http.HandleFunc("/sync-policy", syncPolicyHandler)

	log.Fatal(http.ListenAndServe(":80", nil))
}

// setLogLevel reads desired logging level from the LOG_LEVEL env variable and sets it. Possible
// options are debug, info, warn, error, and fatal. Defaults to logrus's default.
func setLogLevel() {
	var envLogLevel string

	if v := os.Getenv(EnvLogLevel); v != "" {
		envLogLevel = v
	}

	switch envLogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	}
}

func handleError(w http.ResponseWriter, err error, status int) {
	log.WithFields(log.Fields{"status": status}).Error(err.Error())
	http.Error(w, err.Error(), status)
}

func syncKubernetesRoleHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Request to /sync-kubernetes-role")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	request := &VaultKubernetesRoleRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	response := &SyncResponse{}
	response.Status.Succeeded = false
	role := request.Parent.Spec
	v, err := vault.New()
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	if request.Finalizing {
		log.Infof("Deleting Kubernetes Role %s", role.Name)
		if err := v.DeleteKubernetesRole(role); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		response.Status.Succeeded = true
		response.Finalized = true

	} else {
		log.Infof("Creating/updating Kubernetes Role %s", role.Name)
		if err := v.UpdateKubernetesRole(role); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		response.Status.Succeeded = true
	}

	body, err = json.Marshal(&response)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func syncPolicyHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Request to /sync-policy")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	request := &VaultPolicyRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	response := &SyncResponse{}
	response.Status.Succeeded = false
	policy := request.Parent.Spec
	v, err := vault.New()
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	if request.Finalizing {
		log.Infof("Deleting Policy %s", policy.Name)
		if err := v.DeletePolicy(policy); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		response.Status.Succeeded = true
		response.Finalized = true

	} else {
		log.Infof("Creating/updating Policy %s", policy.Name)
		if err := v.UpdatePolicy(policy); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		response.Status.Succeeded = true
	}

	body, err = json.Marshal(&response)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
