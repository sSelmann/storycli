package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of Story and Story-Geth services",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd) // rootCmd, ana komutunuzun adı olabilir. Projenize göre değiştirin.
}

// runStatus executes the status command logic
func runStatus(cmd *cobra.Command, args []string) error {
	printInfo("Checking Story and Geth services exist...")

	storyExists, err := checkServiceExists("story")
	if err != nil {
		return err
	}

	gethExists, err := checkServiceExists("story-geth")
	if err != nil {
		return err
	}

	if !storyExists && !gethExists {
		printWarning("Neither 'story' nor 'story-geth' services are installed.")
		return nil
	}

	if storyExists {
		printInfo("Fetching status for 'story' service...")
		err = displayServiceStatus("story")
		if err != nil {
			// Hatanın devam etmesine gerek yok, ancak fonksiyonun çıktısını görmek için
			// zaten displayServiceStatus fonksiyonu çıktıyı gösteriyor
			// İsteğe bağlı olarak burada işlemlere devam edebilirsiniz
		}
	} else {
		printWarning("'story' service is not installed.")
	}

	if gethExists {
		printInfo("Fetching status for 'story-geth' service...")
		err = displayServiceStatus("story-geth")
		if err != nil {
			// Aynı şekilde, hatanın devam etmesine gerek yok
		}
	} else {
		printWarning("'story-geth' service is not installed.")
	}

	return nil
}

// checkServiceExists checks if a given systemd service exists
func checkServiceExists(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Eğer komut çalıştırılamazsa, servisin mevcut olmadığını varsayabiliriz
		return false, nil
	}

	// Çıktıda servisin adı olup olmadığını kontrol et
	// "story.service" satırını arıyoruz
	return bytes.Contains(out.Bytes(), []byte(serviceName+".service")), nil
}

// displayServiceStatus displays the status of a given systemd service
func displayServiceStatus(serviceName string) error {
	cmdStr := fmt.Sprintf("systemctl status %s --no-pager -fo cat -n 3", serviceName)
	cmd := exec.Command("bash", "-c", cmdStr)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Servis durumu ne olursa olsun çıktıyı göster
	if out.Len() > 0 {
		fmt.Println(out.String())
	}
	if stderr.Len() > 0 {
		fmt.Println("Stderr:", stderr.String())
	}

	// Hata kontrolü
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// *exec.ExitError türündeyse, komut çalıştı ancak çıkış kodu sıfır değil
			// Bu durumda, hizmetin durumu hakkında bilgi aldık, uyarı göstermemize gerek yok
			return nil
		} else {
			// Diğer türde hatalar için uyarı göster
			printWarning(fmt.Sprintf("Failed to get status for %s service.", serviceName))
			return err
		}
	}

	return nil
}
