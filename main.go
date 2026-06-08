package main

import (
	"flag"
	"fmt"
	"runtime"
)

var version string

const goArch string = runtime.GOARCH
const goOS string = runtime.GOOS

var aliases aliasConfig

func main() {

	showVersion := false
	update := false
	dryupd := false
	switchCloud := false
	editMode := false
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&update, "u", false, "Update azs")
	flag.BoolVar(&dryupd, "U", false, "Dry-run update azs")
	flag.BoolVar(&switchCloud, "c", false, "Switch Cloud")
	flag.BoolVar(&editMode, "e", false, "Edit Subscription")
	flag.Parse()

	aliases = newAliases()
	switch {
	case switchCloud:
		clouds := getClouds()
		c := selectClouds(clouds)
		setCloud(c)
		return
	case editMode:
		subscriptions := getSubscriptions()
		s := selectSubscriptions(subscriptions)
		editSubscription(s)
	case showVersion:
		fmt.Println(version, goOS, goArch)
		return
	case update || dryupd:
		updateAzs(dryupd)
		return
	default:
		subscriptions := getSubscriptions()
		s := selectSubscriptions(subscriptions)
		setSubscription(s)
	}
}
