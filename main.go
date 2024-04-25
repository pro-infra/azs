package main

import (
	"flag"
	"fmt"
	"runtime"
)

var version string

const GOARCH string = runtime.GOARCH
const GOOS string = runtime.GOOS

var aliases aliasConfig

func main() {

	showVersion := false
	update := false
	dryupd := false
	switchCloud := false
	editSubscription := false
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&update, "u", false, "Update azs")
	flag.BoolVar(&dryupd, "U", false, "Dry-run update azs")
	flag.BoolVar(&switchCloud, "c", false, "Switch Cloud")
	flag.BoolVar(&editSubscription, "e", false, "Edit Subscription")
	flag.Parse()

	aliases = newAliases()
	switch {
	case switchCloud:
		clouds := getClouds()
		c := select_clouds(clouds)
		set_cloud(c)
		return
	case editSubscription:
		subscriptions := getSubscriptions()
		s := select_subscriptions(subscriptions)
		edit_subscription(s)
	case showVersion:
		fmt.Println(version, GOOS, GOARCH)
		return
	case update || dryupd:
		updateazs(dryupd)
		return
	default:
		subscriptions := getSubscriptions()
		s := select_subscriptions(subscriptions)
		set_subscription(s)
	}
}
