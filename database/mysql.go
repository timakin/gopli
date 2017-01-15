package database

import (
	"bytes"
	"fmt"
	. "github.com/timakin/gopli/constants"
	. "github.com/timakin/gopli/lib"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
)

type MySQLFetcher DBConnector
type MySQLInserter DBConnector

func (fetcher *MySQLFetcher) Fetch() error {
	log.Print("[Fetch] fetching the list of tables...")
	session, err := fetcher.SSHClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var listTableStdoutBuf bytes.Buffer
	session.Stdout = &listTableStdoutBuf
	listTableCmd := fmt.Sprintf(SHOW_TABLES_CMD_FORMAT, fetcher.Name, fetcher.User, fetcher.Password)
	err = session.Run(listTableCmd)

	if err := os.MkdirAll(TMP_DIR_PATH, 0777); err != nil {
		return err
	}

	tableListSavePath := TMP_DIR_PATH + "/table_list.txt"
	ioutil.WriteFile(tableListSavePath, listTableStdoutBuf.Bytes(), os.ModePerm)
	log.Print("[Fetch] completed fetching the list of tables")

	log.Print("\t[Fetch] start to fetch table contents...")
	tables, err := ReadLines(tableListSavePath)
	if err != nil {
		return err
	}

	sem := make(chan int, MaxFetchSession)
	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			sem <- 1
			defer wg.Done()
			defer func() { <-sem }()
			session, err := fetcher.SSHClient.NewSession()
			if err != nil {
				panic(err)
			}
			defer session.Close()

			var fetchResult bytes.Buffer
			session.Stdout = &fetchResult
			fetchRowsCmd := fmt.Sprintf(SELECT_TABLES_CMD_FORMAT, fetcher.User, fetcher.Password, fetcher.Name, table)
			log.Print("\t\t[Fetch] fetching " + table)
			err = session.Run(fetchRowsCmd)
			if err != nil {
				panic(err)
			}
			dumpSavePath := TMP_DIR_PATH + "/" + table + ".txt"
			ioutil.WriteFile(dumpSavePath, fetchResult.Bytes(), os.ModePerm)
			log.Print("\t\t[Fetch] completed fetcing " + table)
		}(table)
	}
	wg.Wait()
	log.Print("\t[Fetch] completed fetching all tables")
	return nil
}

func (inserter *MySQLInserter) Clean() error {
	log.Print("[Delete] deleting existing tables...")
	var tables []string
	tableListSavePath := TMP_DIR_PATH + "/table_list.txt"
	tables, err := ReadLines(tableListSavePath)
	if err != nil {
		return err
	}

	sem := make(chan int, 5)
	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			sem <- 1
			defer wg.Done()
			defer func() { <-sem }()

			log.Print("\t[Delete] deleting " + table)

			if inserter.Host == "localhost" || inserter.Host == "127.0.0.1" {
				var CLEAN_TABLES_CMD_FORMAT *exec.Cmd
				query := fmt.Sprintf(DELETE_TABLE_QUERY_FORMAT, inserter.Name, table)
				userOption := "-u" + inserter.User
				executeOption := "--execute=" + query
				var passwordOption string
				if len(inserter.Password) > 0 {
					passwordOption = "-p" + inserter.Password
				} else {
					passwordOption = ""
				}
				CLEAN_TABLES_CMD_FORMAT = exec.Command("mysql", userOption, passwordOption, executeOption)

				err := CLEAN_TABLES_CMD_FORMAT.Run()
				if err != nil {
					panic(err)
				}
			} else {
				var CLEAN_TABLES_CMD_FORMAT string
				if len(inserter.Password) > 0 {
					CLEAN_TABLES_CMD_FORMAT = fmt.Sprintf(CLEAN_TABLES_CMD_FORMAT, inserter.User, inserter.Password, inserter.Name, table)
				} else {
					CLEAN_TABLES_CMD_FORMAT = fmt.Sprintf(CLEAN_TABLES_CMD_FORMAT_WITHOUT_PASSPHRASE, inserter.User, inserter.Name, table)
				}

				var CleantdoutBuf bytes.Buffer
				session, err := inserter.SSHClient.NewSession()
				if err != nil {
					panic(err)
				}
				defer session.Close()
				session.Stdout = &CleantdoutBuf
				err = session.Run(CLEAN_TABLES_CMD_FORMAT)
				if err != nil {
					panic(err)
				}
			}
		}(table)
	}
	wg.Wait()
	log.Print("[Delete] completed deleting tables")
	return nil
}

func (inserter *MySQLInserter) Insert() error {
	log.Print("[Load Infile] start to send fetched contents...")
	var tables []string
	tableListSavePath := TMP_DIR_PATH + "/table_list.txt"
	tables, err := ReadLines(tableListSavePath)
	if err != nil {
		return err
	}
	sem := make(chan int, MaxLoadInfileSession)
	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			sem <- 1
			defer wg.Done()
			defer func() { <-sem }()
			fetchedTableFile := TMP_DIR_PATH + "/" + table + ".txt"
			query := fmt.Sprintf(LOAD_INFILE_QUERY_FORMAT, fetchedTableFile, inserter.Name, table)
			var passwordOption string
			if len(inserter.Password) > 0 {
				passwordOption = fmt.Sprintf("-p%s", inserter.Password)
			} else {
				passwordOption = ""
			}
			log.Print("\t[Load Infile] start to send the contents inside of " + table)
			var cmd *exec.Cmd
			if inserter.Host == "localhost" || inserter.Host == "127.0.0.1" {
				cmd = exec.Command("mysql", "-u"+inserter.User, passwordOption, "--enable-local-infile", "--execute="+query)
			} else {
				cmd = exec.Command("mysql", "-u"+inserter.User, passwordOption, "-h"+inserter.Host, "--enable-local-infile", "--execute="+query)
			}
			err := cmd.Run()
			if err != nil {
				panic(err)
			}
			log.Print("\t[Load Infile] completed sending the contents inside of " + table)
		}(table)
		wg.Wait()
	}
	log.Print("[Load Infile] completed sending fetched contents")
	log.Print("[Finished] All tasks finished")
	return nil
}
