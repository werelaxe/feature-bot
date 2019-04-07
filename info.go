package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type User struct {
	Name     string
	Count    int
	Nickname string
}

type Info struct {
	Users        map[int]User
	SpecialHuman int
	Feature      string
	LastSpinTime int64
}

type InfoKeeper struct {
	Path string
}

func (ik *InfoKeeper) Read(chatId int64) (*Info, error) {
	var info Info
	data, err := ioutil.ReadFile(path.Join(ik.Path, fmt.Sprint(chatId)))
	if err != nil {
		return nil, errors.New("can not read info: " + err.Error())
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, errors.New("can not read info: " + err.Error())
	}
	return &info, nil
}

func (ik *InfoKeeper) Write(chatId int64, info *Info) error {
	data, err := json.Marshal(*info)
	if err != nil {
		return errors.New("can not write info: " + err.Error())
	}
	if err := ioutil.WriteFile(path.Join(ik.Path, fmt.Sprint(chatId)), data, 0644); err != nil {
		return errors.New("can not write info: " + err.Error())
	}
	return nil
}

func (ik *InfoKeeper) IsChatJoined(chatId int64) bool {
	if _, err := os.Stat(path.Join(ik.Path, fmt.Sprint(chatId))); os.IsNotExist(err) {
		return false
	}
	return true
}

func (ik *InfoKeeper) GetChats() ([]int64, error) {
	files, err := ioutil.ReadDir(ik.Path)
	if err != nil {
		return nil, errors.New("can not get chats: " + err.Error())
	}
	var chats []int64
	for _, file := range files {
		val, err := strconv.ParseInt(file.Name(), 10, 64)
		if err == nil {
			chats = append(chats, val)
		}
	}
	return chats, nil
}

func (ik *InfoKeeper) Init(path string) {
	ik.Path = path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0777); err != nil {
			panic(err)
		}
	}
}
