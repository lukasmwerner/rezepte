package models

type Recipe struct {
	Title    string `yaml:"title"`
	Source   string `yaml:"source"`
	Serves   string `yaml:"serves"`
	Time     string `yaml:"time"`
	Contents string
}
