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

type Linode struct {
	ID     int `json:"LINODEID"`
	Status int `json:"STATUS"`
}

func (l Linode) Running() bool {
	return l.Status == 1
}

func LinodeList(ctx context.Context, apiKey string, linodeId int) ([]Linode, error) {
	params := make(map[string]string)
	if linodeId != 0 {
		params["LinodeID"] = strconv.Itoa(linodeId)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.list", params)
	if err != nil {
		return nil, err
	}
	var parsedData []Linode
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

func LinodeCreate(ctx context.Context, apiKey string, datacenterId, planId, paymentTerm int) (linodeId int, err error) {
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

func LinodeDiskCreateFromDistribution(ctx context.Context, apiKey string, linodeId, distroId int, label string, size int, rootPass, rootSSHKey string) (diskId int, jobId int, err error) {
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
		return 0, 0, err
	}
	var parsedData struct {
		DiskID int `json:"DiskID"`
		JobID  int `json:"JobID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, 0, err
	}
	return parsedData.DiskID, parsedData.JobID, nil
}

func LinodeDiskImagize(ctx context.Context, apiKey string, linodeId, diskId int, description, label string) (imageId int, jobId int, err error) {
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
		return 0, 0, err
	}
	var parsedData struct {
		JobID   int `json:"JobID"`
		ImageID int `json:"ImageID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, 0, err
	}
	return parsedData.ImageID, parsedData.JobID, nil
}

func LinodeDelete(ctx context.Context, apiKey string, linodeId int, skipChecks bool) error {
	params := map[string]string{
		"LinodeID":   strconv.Itoa(linodeId),
		"skipChecks": strconv.FormatBool(skipChecks),
	}
	_, err := makeLinodeRequest(ctx, apiKey, "linode.delete", params)
	return err
}

func LinodeDiskDelete(ctx context.Context, apiKey string, linodeId, diskId int) (int, error) {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
		"DiskID":   strconv.Itoa(diskId),
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.disk.delete", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		JobID int `json:"JobID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.JobID, nil
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

type Kernel struct {
	ID      int       `json:"KERNELID"`
	Label   string    `json:"LABEL"`
	IsXEN   LinodeInt `json:"ISXEN"`
	IsKVM   LinodeInt `json:"ISKVM"`
	IsPVOPS LinodeInt `json:"ISPVOPS"`
}

func AvailKernels(ctx context.Context, apiKey string) ([]Kernel, error) {
	data, err := makeLinodeRequest(ctx, apiKey, "avail.kernels", nil)
	if err != nil {
		return nil, err
	}
	var parsedData []Kernel
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

type LinodeInt int

func (i *LinodeInt) UnmarshalJSON(text []byte) error {
	str := string(text)
	if str == `""` {
		*i = 0
	} else {
		if x, err := strconv.Atoi(str); err != nil {
			return err
		} else {
			*i = LinodeInt(x)
		}
	}
	return nil
}

type Job struct {
	ID             int       `json:"JOBID"`
	LinodeID       int       `json:"LINODEID"`
	EnteredDate    string    `json:"ENTERED_DT"`
	HostStartDate  string    `json:"HOST_START_DT"`
	HostFinishDate string    `json:"HOST_FINISH_DT"`
	Action         string    `json:"ACTION"`
	Label          string    `json:"LABEL"`
	Duration       LinodeInt `json:"DURATION"`
	HostMessage    string    `json:"HOST_MESSAGE"`
	HostSuccess    LinodeInt `json:"HOST_SUCCESS"`
}

func LinodeJobList(ctx context.Context, apiKey string, linodeId, jobId int, pendingOnly bool) ([]Job, error) {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
	}
	if jobId != 0 {
		params["JobID"] = strconv.Itoa(jobId)
	}
	if pendingOnly {
		params["pendingOnly"] = "1"
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.job.list", params)
	if err != nil {
		return nil, err
	}
	var parsedData []Job
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

type LinodeIP struct {
	ID       int    `json:"IPADDRESSID"`
	LinodeID int    `json:"LINODEID"`
	IsPublic int16  `json:"ISPUBLIC"`
	Address  string `json:"IPADDRESS"`
	RDNSName string `json:"RDNS_NAME"`
}

func LinodeIPList(ctx context.Context, apiKey string, linodeId, ipAddressId int) ([]LinodeIP, error) {
	params := make(map[string]string)
	if linodeId != 0 {
		params["LinodeID"] = strconv.Itoa(linodeId)
	}
	if ipAddressId != 0 {
		params["IPAddressID"] = strconv.Itoa(ipAddressId)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.ip.list", params)
	if err != nil {
		return nil, err
	}
	var parsedData []LinodeIP
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

func LinodeBoot(ctx context.Context, apiKey string, linodeId, configId int) (int, error) {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
	}
	if configId != 0 {
		params["ConfigID"] = strconv.Itoa(configId)
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.boot", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		JobID int `json:"JobID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.JobID, nil
}

func LinodeShutdown(ctx context.Context, apiKey string, linodeId int) (int, error) {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.shutdown", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		JobID int `json:"JobID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.JobID, nil
}

func LinodeConfigCreate(ctx context.Context, apiKey string, linodeId, diskId, kernelId int, label string) (int, error) {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
		"KernelID": strconv.Itoa(kernelId),
		"DiskList": strconv.Itoa(diskId),
		"Label":    label,
	}
	data, err := makeLinodeRequest(ctx, apiKey, "linode.config.create", params)
	if err != nil {
		return 0, err
	}
	var parsedData struct {
		ConfigID int `json:"ConfigID"`
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return 0, err
	}
	return parsedData.ConfigID, nil
}

func LinodeConfigDelete(ctx context.Context, apiKey string, linodeId, configId int) error {
	params := map[string]string{
		"LinodeID": strconv.Itoa(linodeId),
		"ConfigID": strconv.Itoa(configId),
	}
	_, err := makeLinodeRequest(ctx, apiKey, "linode.config.delete", params)
	return err
}

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
