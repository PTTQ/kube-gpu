/**
 * Copyright (2021, ) Institute of Software, Chinese Academy of Sciences
 **/

package node_daemon

import (
	"bufio"
	"fmt"
	jsonObj "github.com/kubesys/client-go/pkg/json"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getArchFamily(computeMajor, computeMinor int) string {
	switch computeMajor {
	case 1:
		return "Tesla"
	case 2:
		return "Fermi"
	case 3:
		return "Kepler"
	case 5:
		return "Maxwell"
	case 6:
		return "Pascal"
	case 7:
		if computeMinor < 5 {
			return "volta"
		}
		return "Turing"
	case 8:
		return "Ampere"
	}
	return "Unknown"
}

func getCgroupPath(pod *jsonObj.JsonObject, containerID string) (string, error) {
	meta := pod.GetJsonObject("metadata")
	podUID, err := meta.GetString("uid")
	if err != nil {
		return "", err
	}
	status := pod.GetJsonObject("status")
	qosClass, err := status.GetString("qosClass")
	if err != nil {
		return "", err
	}

	podUID = strings.Replace(podUID, "-", "_", -1) + ".slice"
	path := "kubepods.slice"
	switch qosClass {
	case PodQOSGuaranteed:
		path = filepath.Join(path, "kubepods-guaranteed.slice")
		podUID = "kubepods-guaranteed-pod" + podUID
	case PodQOSBurstable:
		path = filepath.Join(path, "kubepods-burstable.slice")
		podUID = "kubepods-burstable-pod" + podUID
	case PodQOSBestEffort:
		path = filepath.Join(path, "kubepods-besteffort.slice")
		podUID = "kubepods-besteffort-pod" + podUID
	}

	path = filepath.Join(path, podUID)
	return fmt.Sprintf("%s/docker-%s.scope", path, containerID), nil
}

func readProcsFile(file string) ([]int, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Can't read %s, %s.", file, err)
		return nil, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	pids := make([]int, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if pid, err := strconv.Atoi(line); err == nil {
			pids = append(pids, pid)
		}
	}

	log.Infof("Read from %s, pids: %v.", file, pids)
	return pids, nil
}
