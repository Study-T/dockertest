package service

// TrackEvent represents a single tracking event.
type TrackEvent struct {
	ProcessTime     string `json:"process_time"`
	ProcessUTCTime  string `json:"process_utc_time"`
	ProcessContent  string `json:"process_content"`
	ProcessCountry  string `json:"process_country"`
	ProcessProvince string `json:"process_province"`
	ProcessCity     string `json:"process_city"`
	ProcessLocation string `json:"process_location"`
	TrackNodeCode   string `json:"track_node_code"`
	TrackNodeDesc   string `json:"track_node_description"`
	PodURL          string `json:"pod_url"`
}

// TrackInfo contains detailed tracking information.
type TrackInfo struct {
	WaybillNumber                 string       `json:"运单号"`
	TrackingNumber                string       `json:"tracking_number"`
	CustomerOrderNumber           string       `json:"customer_order_number"`
	ProductCode                   string       `json:"product_code"`
	ProductName                   string       `json:"product_name"`
	ChannelCode                   string       `json:"channel_code"`
	CheckInTime                   string       `json:"check_in_time"`
	CheckOutTime                  string       `json:"check_out_time"`
	PickUpTime                    string       `json:"pick_up_time"`
	CustomerCode                  string       `json:"customer_code"`
	OriginCode                    string       `json:"origin_code"`
	DestinationCode               string       `json:"destination_code"`
	PostalCode                    string       `json:"postal_code"`
	ActualWeight                  float64      `json:"actual_weight"`
	IntervalDay                   int          `json:"interval_day"`
	IntervalWorkDay               int          `json:"interval_work_day"`
	LastMileSite                  string       `json:"last_mile_site"`
	LastMileName                  string       `json:"last_mile_name"`
	PhoneNumber                   string       `json:"phone_number"`
	TrackEvents                   []TrackEvent `json:"track_events"`
	PodURL                        string       `json:"pod_url"`
	PodURLs                       []string     `json:"pod_urls"`
	IsSignature                   bool         `json:"IsSignature"`
	SignatureURLs                 []string     `json:"SignatureUrls"`
	EstimatedDeliveryToDateZone   string       `json:"EstimatedDeliveryToDateZone"`
	EstimatedDeliveryFromDateZone string       `json:"EstimatedDeliveryFromDateZone"`
}

// TrackingResult is the query result DTO.
type TrackingResult struct {
	OrderNumber   string    `json:"order_number"`
	PackageStatus string    `json:"package_status"`
	TrackInfo     TrackInfo `json:"track_Info"`
}
