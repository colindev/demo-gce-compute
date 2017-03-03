package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	compute "google.golang.org/api/compute/v1"

	"golang.org/x/oauth2/google"
)

var (
	addr string
	ctx  = context.Background()
)

type key int

var computeServiceKey key

func SetComputeService(service *compute.Service) {
	ctx = context.WithValue(ctx, computeServiceKey, service)
}
func GetComputeService() (*compute.Service, bool) {
	service, ok := ctx.Value(computeServiceKey).(*compute.Service)
	return service, ok
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

	http.Handle("/", http.FileServer(http.Dir("./views")))
	http.HandleFunc("/admin/api/compute/images", listDebianImages)
	http.HandleFunc("/admin/api/compute/instances", listComputeInstances)
	http.HandleFunc("/admin/api/compute/instances/insert", insertComputeInstance)
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

	project := r.FormValue("project")
	name := r.FormValue("name")
	zone := r.FormValue("zone")
	machineType := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/machineTypes/%s", project, zone, r.FormValue("machine_type"))
	image := r.FormValue("image")
	imageURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/%s", image)
	network := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, "default")

	startupScript, err := ioutil.ReadFile("./scripts/install-wp-lamp.sh")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	script := string(startupScript)

	instance := &compute.Instance{
		Name:        name,
		Description: "post via golang server",
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "demo",
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

	op, err := service.Instances.Insert(project, zone, instance).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, op)
}
