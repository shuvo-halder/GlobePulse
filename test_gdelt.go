package main
import (
	"fmt"
	"io/ioutil"
	"net/http"
)
func main() {
	resp, err := http.Get("https://api.gdeltproject.org/api/v2/doc/doc?query=earthquake&mode=artlist&maxrecords=5&format=json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
