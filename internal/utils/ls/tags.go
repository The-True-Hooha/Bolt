package lscmd

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

)


func DoSomething() {
	fmt.Println("hello")
}

var tagFile = filepath.Join(os.Getenv("HOME"), ".bolt_tags.json")

type FileTags map[string][]string

func loadSystemTags()(FileTags, error){
	data, err := os.ReadFile(tagFile)
	if os.IsNotExist(err){
		return make(FileTags), nil
	}else if err != nil {
		return nil, err
	}
	var tags FileTags
	err = json.Unmarshal(data, &tags)
	return tags, err
}

func GetFileTags(filename string) ([]string, error){
	tags, err := loadSystemTags()
	if err != nil {
		return nil, err
	}
	return tags[filename], nil
}

func AddFileTags(filename string, tag string) error{

	tags, err := loadSystemTags()
	if err != nil{
		return err
	}
	tags[filename] = append(tags[filename], tag)

	return saveSystemTags(tags)
}

func saveSystemTags(tags FileTags) error {
	data, err := json.Marshal(tags)
	if err != nil{
		return err
	}
	return os.WriteFile(tagFile, data, 0644)
}

func RemoveFileTags(filename string, tag string) error{
	tags, err := loadSystemTags()
	if err != nil{
		return nil
	}
	for i, t := range tags[filename] {
		if t == tag {
			tags[filename] = append(tags[filename][:i], tags[filename][i+1:]...)
			break
		}
	}
	return saveSystemTags(tags)
}




