package cfg

var ConfigVar = Config{}

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	} `json:"database"`

	Logging struct {
		OutFile string `json:"out_file"`
	} `json:"logging"`

	SMTP struct {
		SenderEmail  string `json:"sender_email"`
		SenderPasswd string `json:"sender_password"`
	} `json:"SMTP"`
}
