package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Result struct {
	LastSeen time.Time
	Changed  time.Time
	Active   bool
	IP       string
	Name     string
	Widget   *widget.Button
}

var ip = binding.NewString()

func main() {
	app := app.New()

	mainWindow := app.NewWindow("Pinger")
	// ic, err := fyne.LoadResourceFromPath("./logo.png")
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// mainWindow.SetIcon(ic)

	// mainWindow.CenterOnScreen()
	mainWindow.Resize(fyne.NewSize(300, 400))
	mainWindow.Show()

	// Whilst this is never seen, it ensures the widget is wider
	nets := GetIPv4NonLocalInterfaces()
	wSelectNet := widget.NewSelect(nets, func(s string) {
		ip.Set(s)
	})
	wSelectNet.PlaceHolder = "Select network"

	top := container.NewHBox(
		wSelectNet,
		layout.NewSpacer(),
		widget.NewButton("Scan", func() {
			Scan(mainWindow)
		}),
	)

	body := container.NewHScroll(widget.NewLabelWithData(ip))

	content := container.NewBorder(
		top,
		nil,
		nil,
		nil,
		body,
	)

	mainWindow.SetContent(content)

	app.Run()
}

// Scan starts a scan, cancelling a current scan if applicable
func Scan(w fyne.Window) {
	ip, err := ip.Get()
	if err != nil {
		dialog.NewInformation("Info", "Error: "+err.Error(), w).Show()
		return
	}
	if ip == "" {
		dialog.NewInformation("Info", "Not IP", w).Show()
		return
	}

	go func() {
		_, err := Hosts(ip)

		if err != nil {
			dialog.NewInformation("Info", "Error: "+err.Error(), w).Show()
			return
		}
	}()

	// go func() {
	// 	ipList, err := Hosts(a.ip)
	// 	if err != nil {
	// 		a.wStatus.SetText("Error: " + err.Error())
	// 		return
	// 	}

	// 	if a.cancelScan != nil {
	// 		a.cancelScan()
	// 	}
	// 	var ctx context.Context

	// 	ctx, a.cancelScan = context.WithCancel(context.Background())
	// 	a.wStatus.SetText("Scanning: " + a.ip)
	// 	results, err := a.startScan(ctx, ipList)
	// 	a.wStatus.SetText("")
	// 	if err != nil {
	// 		if errors.Is(err, context.Canceled) {
	// 			return
	// 		}
	// 		a.wStatus.SetText("Error: " + err.Error())
	// 		return
	// 	}

	// 	now := time.Now()
	// 	a.Lock()
	// 	// Add/update results
	// 	for i, item := range results {
	// 		// nowi is the time the scan starts incremented by a ms to ensure the results are sorted neatly
	// 		nowi := now.Add(time.Millisecond * time.Duration(i))
	// 		if r, ok := a.results[item.IP.String()]; ok { // Already in results
	// 			if !r.Active {
	// 				r.Changed = nowi
	// 			}
	// 			r.LastSeen = nowi
	// 			r.Active = true
	// 			r.Widget.active = true
	// 		} else { // Add to cache
	// 			a.results[item.IP.String()] = &Result{
	// 				Active:   true,
	// 				LastSeen: nowi,
	// 				Changed:  nowi,
	// 				IP:       item.IP.String(),
	// 				// Name:     item.Name,
	// 				Widget: newIPWidget(item.IP.String()),
	// 			}
	// 		}
	// 	}
	// 	// Mark results that are no longer pinging
	// 	for ip, item := range a.results {
	// 		exists := false
	// 		for _, r := range results {
	// 			if r.IP.String() == ip {
	// 				exists = true
	// 			}
	// 		}
	// 		if !exists {
	// 			if item.Active {
	// 				item.Active = false
	// 				item.Changed = now.Add(-time.Second) // sorts inactive below active
	// 				item.Widget.active = false
	// 			}
	// 		}
	// 	}

	// 	// Create a copy of the results map as an array to sort
	// 	arr := make([]*Result, 0)
	// 	for _, r := range a.results {
	// 		arr = append(arr, r)
	// 	}
	// 	sort.Slice(arr, func(i, j int) bool {
	// 		return arr[i].Changed.After(arr[j].Changed)
	// 	})

	// 	// Copy the sorted results array to the fyne grid widget
	// 	a.cResults.Objects = make([]fyne.CanvasObject, len(arr))
	// 	for i, r := range arr {
	// 		a.cResults.Objects[i] = r.Widget
	// 	}
	// 	autoscan := a.autoscan
	// 	a.Unlock()

	// 	a.cResults.Refresh()
	// 	if autoscan {
	// 		// Have a break before starting the next scan
	// 		time.Sleep(time.Millisecond * 100)
	// 		a.Scan()
	// 	}
	// }()
}

func Hosts(cidr string) ([]net.IP, error) {
	// convert string to IPNet struct
	_, ipv4Net, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	// convert IPNet struct mask and address to uint32
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)

	var ips []net.IP

	// loop through addresses as uint32
	for i := start + 1; i <= finish-1; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)

		log.Println(ip.String())
		log.Println(ip.DefaultMask())
		log.Println(ip.IsLoopback())
		log.Println(ip.IsGlobalUnicast())
		log.Println(ip.IsMulticast())

		ips = append(ips, ip)
	}
	return ips, nil
}
