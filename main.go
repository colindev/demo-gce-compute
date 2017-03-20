package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"

	"github.com/colindev/wshub"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"

	"golang.org/x/oauth2/google"
)

var (
	ctx = context.Background()
)

type key int

var (
	computeServiceKey key = 1
	wshubKey          key = 2
	env                   = struct {
		Addr           string
		ProjectID      string
		BasePath       string
		SelfInternalIP string
	}{
		ProjectID:      "gcetest-156204",
		SelfInternalIP: "127.0.0.1",
	}
)

// SetComputeService ...
func SetComputeService(service *compute.Service) {
	ctx = context.WithValue(ctx, computeServiceKey, service)
}

// GetComputeService ...
func GetComputeService() (*compute.Service, bool) {
	service, ok := ctx.Value(computeServiceKey).(*compute.Service)
	return service, ok
}

// SetWSHub ...
func SetWSHub(hub *wshub.Hub) {
	ctx = context.WithValue(ctx, wshubKey, hub)
}

// GetWSHub ...
func GetWSHub() (*wshub.Hub, bool) {
	hub, ok := ctx.Value(wshubKey).(*wshub.Hub)
	return hub, ok
}

// ProcessStatus ...
type ProcessStatus struct {
	Active string `json:"active"`
	Items  Items  `json:"items"`
}

// Items ...
type Items map[string]string

func init() {
	log.SetFlags(log.Lshortfile)

	var err error
	if metadata.OnGCE() {
		env.ProjectID, err = metadata.ProjectID()
		if err != nil {
			log.Fatal(err)
		}
		env.SelfInternalIP, err = metadata.InternalIP()
		if err != nil {
			log.Fatal(err)
		}
	}

	basePath, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(basePath + "/public")
	if err != nil {
		basePath = filepath.Dir(os.Args[0])
	}

	env.BasePath = basePath
	flag.StringVar(&env.Addr, "addr", ":80", "http address")

}

func main() {

	flag.Parse()
	log.Printf("%+v", env)

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		log.Fatal(err)
	}

	service, err := compute.New(client)
	if err != nil {
		log.Fatal(err)
	}

	SetComputeService(service)

	hub := wshub.New()
	go hub.Run()
	SetWSHub(hub)

	// views
	fsServer := http.FileServer(http.Dir(env.BasePath + "/public"))

	tpls := map[string]*template.Template{}
	(func() {
		mainTpl := env.BasePath + "/templates/main.tpl"
		tplDir := env.BasePath + "/templates/contents"
		fs, err := ioutil.ReadDir(tplDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range fs {
			name := f.Name()
			if strings.HasSuffix(name, ".tpl") {
				tpls[strings.TrimSuffix(name, ".tpl")] = template.Must(template.ParseFiles(mainTpl, tplDir+"/"+name))
			}

		}
	})()
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimLeft(r.URL.Path, "/")
		name = strings.TrimSuffix(name, ".html")
		if name == "" {
			name = "index"
		}
		if tpl, ok := tpls[name]; ok {
			tpl.Execute(w, nil)
			return
		}
		fsServer.ServeHTTP(w, r)
	}))
	// ws
	http.Handle("/ws", hub.Handler())
	http.HandleFunc("/ws-broadcast", broadcast)
	// admin apis
	http.HandleFunc("/admin/api/compute/zones", listZones)
	http.HandleFunc("/admin/api/compute/images", listImages)
	http.HandleFunc("/admin/api/compute/instances", listComputeInstances)
	http.HandleFunc("/admin/api/compute/instance", getComputeInstance)
	http.HandleFunc("/admin/api/compute/instances/insert", insertComputeInstance)
	http.HandleFunc("/admin/api/compute/instances/delete", deleteConputeInstance)
	// run server
	log.Println(http.ListenAndServe(env.Addr, nil))
}

func writeRes(w http.ResponseWriter, v interface{}) {

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func listZones(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
	}
	query.Read(r.URL.Query())

	res, err := service.Zones.List(query["project"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func listImages(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": "centos-cloud",
	}
	query.Read(r.URL.Query())

	res, err := service.Images.List(query["project"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func getComputeInstance(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
		"zone":    "",
		"name":    "",
	}
	query.Read(r.URL.Query())

	res, err := service.Instances.Get(query["project"], query["zone"], query["name"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func listComputeInstances(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
		"zone":    "",
	}
	query.Read(r.URL.Query())

	res, err := service.Instances.List(query["project"], query["zone"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func insertComputeInstance(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	r.ParseForm()
	query := Config{
		"project":        env.ProjectID,
		"zone":           "asia-east1-a",
		"cpu":            "1",    // vCPU
		"memory":         "1024", // MB
		"network":        "default",
		"cloud_image":    "centos-cloud",
		"image":          "centos-7-v20170227",
		"name":           "",
		"startup_script": "",
	}
	query.Read(r.Form)

	machineType := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/machineTypes/custom-%s-%s",
		query["project"],
		query["zone"],
		query["cpu"],
		query["memory"])

	imageURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/images/%s",
		query["cloud_image"],
		query["image"])

	network := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", query["project"], query["network"])

	callbackURL := "http://" + env.SelfInternalIP + "/ws-broadcast"
	script := fmt.Sprintf(`#!/usr/bin/env bash

cat <<EOF > /tmp/startup-script

%s

EOF

chmod +x /tmp/startup-script

gsutil cp gs://demo-compute/installer /tmp/
chmod +x /tmp/installer

CALLBACK_URL=%s /tmp/installer /tmp/startup-script
	
	`,
		query["startup_script"],
		callbackURL,
	)

	instance := &compute.Instance{
		Name:        query["name"],
		Description: "post via golang server",
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "disk-" + query["name"],
					SourceImage: imageURL,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: network,
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: "default",
				Scopes: []string{
					"https://www.googleapis.com/auth/devstorage.read_only",
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "startup-script",
					Value: &script,
				},
				{
					Key:   "callback-url",
					Value: &callbackURL,
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"http-server", "https-server"},
		},
	}

	b, _ := json.MarshalIndent(instance, "", "  ")
	fmt.Println(string(b))

	op, err := service.Instances.Insert(query["project"], query["zone"], instance).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	go func() {

		defer log.Println("quit check instance")
		hub, exists := GetWSHub()
		if !exists {
			return
		}

		for {
			time.Sleep(time.Second)
			inst, err := service.Instances.Get(query["project"], query["zone"], query["name"]).Do()
			if googleapi.IsNotModified(err) {
				hub.Broadcast(ProcessStatus{
					Active: "compute#create",
					Items: Items{
						"not-modified": err.Error(),
					},
				})
			} else if err != nil {
				hub.Broadcast(ProcessStatus{
					Active: "compute#create",
					Items: Items{
						"error": err.Error(),
					},
				})
			}
			hub.Broadcast(ProcessStatus{
				Active: "compute#instance#" + inst.Name,
				Items: Items{
					"status": inst.Status,
				},
			})

			if inst.Status == "RUNNING" {
				hub.Broadcast(ProcessStatus{
					Active: "compute#instance#" + inst.Name,
					Items: Items{
						"network-ip": inst.NetworkInterfaces[0].NetworkIP,
						"nat-ip":     inst.NetworkInterfaces[0].AccessConfigs[0].NatIP,
					},
				})
				return
			}
		}

	}()

	writeRes(w, op)
}

func deleteConputeInstance(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	project := r.FormValue("project")
	zone := r.FormValue("zone")
	name := r.FormValue("name")

	op, err := service.Instances.Delete(project, zone, name).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, op)
}

func broadcast(w http.ResponseWriter, r *http.Request) {

	hub, exists := GetWSHub()
	if !exists {
		http.Error(w, "wshub not found", 500)
		return
	}

	var status ProcessStatus
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hub.Broadcast(status)
}
