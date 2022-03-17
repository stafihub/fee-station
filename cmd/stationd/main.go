// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	_ "fee-station/cmd/stationd/docs"
	"fee-station/dao/migrate"
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/log"
	"fee-station/pkg/utils"
	"fee-station/server"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func _main() error {
	cfg, err := config.Load("conf_station.toml")
	if err != nil {
		fmt.Printf("loadConfig err: %s", err)
		return err
	}
	if cfg.Mode == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}
	log.InitLogFile(cfg.LogFilePath + "/station")
	logrus.Infof("config info: \nlistenAddr: %s\nswapRate: %s\nswapMaxLimit: %s\nswapMinLimit: %s\nlogFilePath: %s\ntokenInfo: %+v\n",
		cfg.ListenAddr, cfg.SwapRate, cfg.SwapMaxLimit, cfg.SwapMinLimit, cfg.LogFilePath, cfg.TokenInfo)

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

	err = migrate.AutoMigrate(db)
	if err != nil {
		logrus.Errorf("dao autoMigrate err: %s", err)
		return err
	}
	//server
	server, err := server.NewServer(cfg, db)
	if err != nil {
		logrus.Errorf("new server err: %s", err)
		return err
	}
	err = server.Start()
	if err != nil {
		logrus.Errorf("server start err: %s", err)
		return err
	}
	defer func() {
		logrus.Infof("shutting down server ...")
		server.Stop()
	}()

	<-ctx.Done()
	return nil
}

// @title feeStation API
// @version 1.0
// @description feeStation api document.

// @contact.name tk
// @contact.email tpkeeper@qq.com

// @host localhost:8083
// @BasePath /feeStation/api
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(40)
	err := _main()
	if err != nil {
		os.Exit(1)
	}
}
