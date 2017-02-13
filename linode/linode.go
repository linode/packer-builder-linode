package linode

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-multierror"
)

type LinodeError struct {
	Code    int    `json:"ERRORCODE"`
	Message string `json:"ERRORMESSAGE"`
}

func (e LinodeError) Error() string {
	return e.Message
}

type LinodeResponse struct {
	ErrorArray []LinodeError   `json:"ERRORARRAY"`
	Action     string          `json:"ACTION"`
	Data       json.RawMessage `json:"DATA"`
}

func (r LinodeResponse) Errors() []error {
	errs := make([]error, len(r.ErrorArray))
	for i := range r.ErrorArray {
		errs[i] = r.ErrorArray[i]
	}
	return errs
}

func LinodeCreate(ctx context.Context, apiKey string, datacenterId, planId, paymentTerm int) (int, error) {
	params := map[string]string{
		"DatacenterID": strconv.Itoa(datacenterId),
		"PlanID":       strconv.Itoa(planId),
	}
	if paymentTerm != 0 {
		params["PaymentTerm"] = strconv.Itoa(paymentTerm)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.create", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		LinodeID int `json:"LinodeID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.LinodeID, nil
}

func LinodeDiskCreateFromDistribution(ctx context.Context, apiKey string, linodeId, distroId int, label string, size int, rootPass, rootSSHKey string) (int, error) {
	params := map[string]string{
		"LinodeID":       strconv.Itoa(linodeId),
		"DistributionID": strconv.Itoa(distroId),
		"Label":          label,
		"Size":           strconv.Itoa(size),
		"rootPass":       rootPass,
	}
	if rootSSHKey != "" {
		params["rootSSHKey"] = rootSSHKey
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.disk.createfromdistribution", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		JobID  int `json:"JobID"`
		DiskID int `json:"DiskID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.DiskID, nil
}

func LinodeDiskImagize(ctx context.Context, apiKey string, linodeId, diskId int, description, label string) (int, error) {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
		"DiskID":   strconv.Itoa(diskId),
	}
	if description != "" {
		params["Description"] = description
	}
	if label != "" {
		params["Label"] = label
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.disk.imagize", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		JobID   int `json:"JobID"`
		ImageID int `json:"ImageID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.ImageID, nil
}

func LinodeDelete(ctx context.Context, apiKey string, linodeId int, skipChecks bool) error {
	params := map[string]string{
		"LinodeID":   strconv.Itoa(linodeId),
		"skipChecks": strconv.FormatBool(skipChecks),
	}
	_, err := makeLinodeRequest(ctx, apiKey, "linode.delete", params)
	return err
}

type Distribution struct {
	ID                  int    `json:"DISTRIBUTIONID"`
	Is64Bit             int16  `json:"IS64BIT"` // 1 or 0
	Label               string `json:"LABEL"`
	MinImageSize        int    `json:"MINIMAGESIZE"`
	CreateDate          string `json:"CREATE_DT"`           // Make this a time.Time?
	RequiresPVOPSKernel int16  `json:"REQUIRESPVOPSKERNEL"` // 1 or 0
}

func AvailDistributions(ctx context.Context, apiKey string, distributionId int) ([]Distribution, error) {
	params := make(map[string]string)
	if distributionId != 0 {
		params["DistributionID"] = strconv.Itoa(distributionId)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "avail.distributions", params)
	if err != nil {
		return nil, err
	}
	var parsedData []Distribution
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

type Datacenter struct {
	ID       int    `json:"DATACENTERID"`
	Location string `json:"LOCATION"`
	Abbr     string `json:"ABBR"`
}

func AvailDatacenters(ctx context.Context, apiKey string) ([]Datacenter, error) {
	data, err := makeLinodeRequest(ctx, apiKey, "avail.datacenters", nil)
	if err != nil {
		return nil, err
	}
	var parsedData []Datacenter
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

type Plan struct {
	ID    int            `json:"PLANID"`
	Cores int            `json:"CORES"`
	Price float32        `json:"PRICE"`
	RAM   int            `json:"RAM"`
	Xfer  int            `json:"XFER"`
	Label string         `json:"LABEL"`
	Avail map[string]int `json:"AVAIL"` // from datacenter to count
}

func AvailPlans(ctx context.Context, apiKey string, planId int) ([]Plan, error) {
	params := make(map[string]string)
	if planId != 0 {
		params["PlanID"] = strconv.Itoa(planId)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "avail.linodeplans", params)
	if err != nil {
		return nil, err
	}
	var parsedData []Plan
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

type Image struct {
	ID          int    `json:"IMAGEID"`
	CreateDate  string `json:"CREATE_DT"`
	LastUsed    string `json:"LAST_USED_DT"`
	Creator     string `json:"CREATOR"`
	Label       string `json:"LABEL"`
	Description string `json:"DESCRIPTION"`
	Filesystem  string `json:"FS_TYPE"`
	IsPublic    int16  `json:"ISPUBLIC"`
	MinSize     int    `json:"MINSIZE"`
	Type        string `json:"TYPE"`
}

func ImageList(ctx context.Context, apiKey string, pending bool, imageId int) ([]Image, error) {
	params := make(map[string]string)
	if pending {
		params["pending"] = "1"
	}
	if imageId != 0 {
		params["ImageID"] = strconv.Itoa(imageId)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "image.list", params)
	if err != nil {
		return nil, err
	}
	var parsedData []Image
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

func ImageDelete(ctx context.Context, apiKey string, imageId int) error {
	params := map[string]string{
		"ImageID": strconv.Itoa(imageId),
	}
	_, err := makeLinodeRequest(ctx, apiKey, "image.delete", params)
	return err
}

// makeLinodeRequest executes a request against Linode's API and returns a
// map corresponding to the response's DATA field. Any HTTP or API errors
// are returned as a slice of errors.
func makeLinodeRequest(ctx context.Context, apiKey, name string, parameters map[string]string) (json.RawMessage, error) {
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("api_action", name)
	for name, value := range parameters {
		params.Add(name, value)
	}

	uri := url.URL{
		Scheme:   "https",
		Host:     "api.linode.com",
		Path:     "/",
		RawQuery: params.Encode(),
	}

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var parsedResponse LinodeResponse
	if err := json.Unmarshal(data, &parsedResponse); err != nil {
		return nil, err
	}
	if errs := parsedResponse.Errors(); len(errs) > 0 {
		return nil, &multierror.Error{Errors: errs}
	}
	return parsedResponse.Data, nil
}
