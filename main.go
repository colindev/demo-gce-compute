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
	"path"
	"strings"

	"github.com/colindev/wshub"

	compute "google.golang.org/api/compute/v1"

	"golang.org/x/oauth2/google"
)

var (
	addr string
	ctx  = context.Background()
)

type key int

var (
	computeServiceKey key = 1
	wshubKey          key = 2
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
	Hostname   string  `json:"hostname"`
	Active     string  `json:"active"`
	Percentage float64 `json:"percentage,omitempty"`
}

func init() {
	flag.StringVar(&addr, "addr", ":80", "http address")
	log.SetFlags(log.Lshortfile)
}

func main() {

	flag.Parse()

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
	viewDir := "./public"
	(func() {
		_, err := os.Stat(viewDir)
		if err == nil {
			return
		}
		viewDir = path.Dir(os.Args[0]) + "/public"
	})()
	fsServer := http.FileServer(http.Dir(viewDir))

	tpls := map[string]*template.Template{}
	(func() {
		mainTpl := path.Dir(viewDir) + "/templates/main.tpl"
		tplDir := path.Dir(viewDir) + "/templates/contents"
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
	http.HandleFunc("/admin/api/compute/images", listDebianImages)
	http.HandleFunc("/admin/api/compute/instances", listComputeInstances)
	http.HandleFunc("/admin/api/compute/instance", getComputeInstance)
	http.HandleFunc("/admin/api/compute/instances/insert", insertComputeInstance)
	http.HandleFunc("/admin/api/compute/instances/delete", deleteConputeInstance)
	// run server
	log.Println(http.ListenAndServe(addr, nil))
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

	project := r.FormValue("project")

	res, err := service.Zones.List(project).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func listDebianImages(w http.ResponseWriter, r *http.Request) {

	service, exists := GetComputeService()
	if !exists {
		http.Error(w, "compute service not found", 500)
		return
	}

	res, err := service.Images.List("debian-cloud").Do()
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

	project := r.FormValue("project")
	zone := r.FormValue("zone")
	name := r.FormValue("name")

	res, err := service.Instances.Get(project, zone, name).Do()
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

	project := r.FormValue("project")
	zone := r.FormValue("zone")

	res, err := service.Instances.List(project, zone).Do()
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
	setting := Config{
		"project":        "gcetest-156204",
		"zone":           "asia-east1-a",
		"cpu":            "1",    // vCPU
		"memory":         "1024", // MB
		"network":        "default",
		"cloud_image":    "centos-cloud",
		"image":          "centos-7-v20170227",
		"name":           "",
		"startup_script": "",
	}
	setting.Read(r.Form)

	machineType := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/machineTypes/custom-%s-%s",
		setting["project"],
		setting["zone"],
		setting["cpu"],
		setting["memory"])

	imageURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/images/%s",
		setting["cloud_image"],
		setting["image"])

	network := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", setting["project"], setting["network"])

	script := setting["startup_script"]

	instance := &compute.Instance{
		Name:        setting["name"],
		Description: "post via golang server",
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "disk-" + setting["name"],
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
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "startup-script",
					Value: &script,
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"http-server", "https-server"},
		},
	}

	b, _ := json.MarshalIndent(instance, "", "  ")
	fmt.Println(string(b))

	op, err := service.Instances.Insert(setting["project"], setting["zone"], instance).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	go func() {
		// check instance done

		// hub.Broadcast(status)
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
