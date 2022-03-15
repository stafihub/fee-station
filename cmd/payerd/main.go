// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/log"
	"fee-station/pkg/utils"
	"fee-station/task/payer"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"github.com/urfave/cli/v2"
)

func _main(ctxCli *cli.Context) error {
	cfg, err := config.Load("conf_payer.toml")
	if err != nil {
		fmt.Printf("loadConfig err: %s", err)
		return err
	}
	log.InitLogFile(cfg.LogFilePath + "/payer")
	logrus.Infof("config info:%+v ", cfg)

	//init db
	db, err := db.NewDB(&db.Config{
		Host:   cfg.Db.Host,
		Port:   cfg.Db.Port,
		User:   cfg.Db.User,
		Pass:   cfg.Db.Pwd,
		DBName: cfg.Db.Name,
		Mode:   cfg.Mode})
	if err != nil {
		logrus.Errorf("db err: %s", err)
		return err
	}
	logrus.Infof("db connect success")

	//interrupt signal
	ctx := utils.ShutdownListener()
	defer func() {
		sqlDb, err := db.DB.DB()
		if err != nil {
			logrus.Errorf("db.DB() err: %s", err)
			return
		}
		logrus.Infof("shutting down the db ...")
		sqlDb.Close()
	}()

	fmt.Printf("Will open stafihub wallet from <%s>. \nPlease ", cfg.KeystorePath)
	key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, cfg.KeystorePath, os.Stdin)
	if err != nil {
		return err
	}
	client, err := hubClient.NewClient(key, cfg.PayerAccount, cfg.GasPrice, cfg.StafiHubEndpoint)
	if err != nil {
		return fmt.Errorf("hubClient.NewClient err: %s", err)
	}

	t := task.NewTask(cfg, db, client)
	err = t.Start()
	if err != nil {
		logrus.Errorf("task start err: %s", err)
		return err
	}
	defer func() {
		logrus.Infof("shutting down task ...")
		t.Stop()
	}()

	<-ctx.Done()
	return nil
}

func main() {
	if err := app.Run(os.Args); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}

var app = cli.NewApp()

var cliFlags = []cli.Flag{
	ConfigPath,
}

// init initializes CLI
func init() {
	app.Action = _main
	app.Copyright = "Copyright 2021 Stafi Protocol Authors"
	app.Name = "payer"
	app.Usage = "payerd"
	app.Authors = []*cli.Author{{Name: "Stafi Protocol 2021"}}
	app.Version = "0.0.1"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{}

	app.Flags = append(app.Flags, cliFlags...)
}
