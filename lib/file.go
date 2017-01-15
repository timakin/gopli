package lib

import (
	"bufio"
	"os"
)

var tableBlackList = [4]string{"ar_internal_metadata", "schema_migrations", "repli_chk", "repli_clock"}

func isInBlackList(table string) bool {
	for _, blackListElem := range tableBlackList {
		if blackListElem == table {
			return true
		}
	}
	return false
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if isInBlackList(scanner.Text()) {
			continue
		}
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
