package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/soniah/gosnmp"
)

/*switch_name 為設備名稱，為了區別為cisco 或是 juniper 的機器，因為其OID在輸出上會有些微不同，
所以要區分開來。port_table，則紀錄著每個 port 現在的連線狀況。*/
var (
	port_no           int
	switch_name       string
	port_table        map[int]string
	mode_table        map[int]string
	vlan_table        map[int]string
	native_vlan_table map[int]string
)

func main() {
	Ip := []string{"192.168.2.1", "192.168.2.3"}
	port_table = make(map[int]string)
	mode_table = make(map[int]string)
	vlan_table = make(map[int]string)
	native_vlan_table = make(map[int]string)
	/*一次進行一個機器*/
	for i := 0; i < len(Ip); i++ {
		/*連線*/
		gosnmp.Default.Target = Ip[i]
		err := gosnmp.Default.Connect()
		if err != nil {
			log.Fatalf("Connect() err:%v", err)
		}
		defer gosnmp.Default.Conn.Close()

		/*要查詢的OID
		1.0.8802.1.1.2.1.3.4		lldpLocSysDesc
		1.0.8802.1.1.2.1.4.1.1.8	lldpRemPortDesc
		1.0.8802.1.1.2.1.4.1.1.9	ldpRemSysName
		*/
		oids := []string{"1.0.8802.1.1.2.1.3.4", "1.0.8802.1.1.2.1.4.1.1.8", "1.0.8802.1.1.2.1.4.1.1.9"}
		for i := 0; i < len(oids); i++ {
			switch i {
			case 0:
				err = gosnmp.Default.Walk(oids[i], printswitchDescr)

				/*初始化 port_table，在每一個新的機器(Ip)都要重製表。*/
				intialize_map()

				if err != nil {
					fmt.Printf("Walk Error: %v\n", err)
					os.Exit(1)
				}
			default:
				err = gosnmp.Default.Walk(oids[i], set_port_table)
				if err != nil {
					fmt.Printf("Walk Error:%v\n", err)
					os.Exit(1)
				}
			}
		}
		/*Native Vlan ID
		Cisco:	1.3.6.1.4.1.9.9.46.1.6.1.1.5 vlanTrunkPortNativeVlan
		Juniper:1.3.6.1.2.1.17.7.1.4.5.1.1	dot1qPvid
		*/
		oids = []string{"1.3.6.1.4.1.9.9.46.1.6.1.1.5", "1.3.6.1.2.1.17.7.1.4.5.1.1"}
		switch switch_name {
		case "Cisco":
			err = gosnmp.Default.Walk(oids[0], vlanTrunkPortNativeVlan)
			if err != nil {
				fmt.Printf("nativeID walk Error:%v\n", err)
				os.Exit(1)
			}
		case "Juniper":
			err = gosnmp.Default.Walk(oids[1], dot1qPvid)
			if err != nil {
				fmt.Printf("nativeID walk Error:%v\n", err)
				os.Exit(1)
			}
		}

		/*vlan access or trunk
		Cisco:1.3.6.1.4.1.9.9.46.1.6.1.1.14 		TrunkPortDynamicStatus
		Juniper:1.3.6.1.4.1.2636.3.40.1.5.1.7.1.5.3 jnxExVlanPortAccessMode*/

		/*Cisco trunk 通過的 vlan，因為 Cisco 設定為 trunk mode 之後，就不會再任一個 vlan 當中出現，與 Juniper 	trunk 同時出現在很多個 vlan 不同故要額外實作一個找出其通過vlan的方法*/

		oids = []string{"1.3.6.1.4.1.9.9.46.1.6.1.1.14", "1.3.6.1.4.1.2636.3.40.1.5.1.7.1.5.3"}
		if switch_name == "Cisco" {
			err = gosnmp.Default.Walk(oids[0], TrunkPortDynamicStatus)
			if err != nil {
				fmt.Printf("access_mode Walk Error:%v\n", err)
				os.Exit(1)
			}
		} else {
			err = gosnmp.Default.Walk(oids[1], jnxExVlanPortAccessMode)
			if err != nil {
				fmt.Printf("access_mode Walk Error:%v\n", err)
				os.Exit(1)
			}

		}
		/*port vlan 查詢，因為 Cisco & Juniper OID 不同，對於 Port 使用的 index 方法也不同，
		必須分開時做，用前面儲存的 switch_name 進行判別式，區分出兩種機器實作，同時若 mode
		為 trunk，cisco 則會在 TrunkPortDynamicStatus 做完
		1.3.6.1.4.1.9.9.68.1.2.2.1.2		vmVlan
		1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3	jnxExVlanPortStatus*/
		oids = []string{"1.3.6.1.4.1.9.9.68.1.2.2.1.2", "1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3"}
		if switch_name == "Cisco" {
			err = gosnmp.Default.Walk(oids[0], vmVlan)
			if err != nil {
				fmt.Printf("vlan_set_table Walk Error:%v\n", err)
				os.Exit(1)
			}
		} else {
			err = gosnmp.Default.Walk(oids[1], jnxExVlanPortStatus)
			if err != nil {
				fmt.Printf("vlan_set_table Walk Error:%v\n", err)
				os.Exit(1)
			}
		}

		printTable()

	}
}
func dot1qPvid(pdu gosnmp.SnmpPDU) error {
	vlan_name := jnxExVlanName(gosnmp.ToBigInt(pdu.Value).String())
	if vlan_name == "vlan1" {
		return nil
	}
	oid := strings.Split(pdu.Name, ".")
	ifindex := dot1dBasePortIfIndex(oid[len(oid)-1])
	port_nu := ifDescr(ifindex)
	native_vlan_table[port_nu] = vlan_name
	return nil
}
func vlanTrunkPortNativeVlan(pdu gosnmp.SnmpPDU) error {
	value := gosnmp.ToBigInt(pdu.Value).String()
	if value == "1" {
		return nil
	}
	oid := strings.Split(pdu.Name, ".")
	port_nu := ifDescr(oid[len(oid)-1])
	native_vlan_table[port_nu] = "vlan" + value
	return nil
}
func jnxExVlanPortAccessMode(pdu gosnmp.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".")
	ifindex := dot1dBasePortIfIndex(oid[len(oid)-1])
	port_nu := ifDescr(ifindex)
	mode := gosnmp.ToBigInt(pdu.Value).String()
	if mode == "1" {
		mode_table[port_nu] = "access"
	} else {
		mode_table[port_nu] = "trunk"
	}
	return nil
}
func TrunkPortDynamicStatus(pdu gosnmp.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".")
	port_nu := ifDescr(oid[len(oid)-1])
	mode := gosnmp.ToBigInt(pdu.Value).String()
	if mode == "2" {
		mode_table[port_nu] = "access"
		return nil
	}
	mode_table[port_nu] = "trunk"

	//查詢 trunk 相關 vlan
	oids := []string{"1.3.6.1.4.1.9.9.46.1.6.1.1.4." + oid[len(oid)-1]}
	result, err := gosnmp.Default.Get(oids)
	if err != nil {
		fmt.Printf("Get TrunkPortDynamicStatus Error:%v\n", err)
		os.Exit(1)
	}
	for _, v := range result.Variables {
		/*將octetString 編碼成 hexadecimal encoding ，共有 32*8個數字
		e.g.
		60 00 00 00 00 00 00 00 00 00 00 00 00 01 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00

		6 in hex to binary is 0110 第一個數字為 0~3 第二個字為 4~7 以此類推
		故 0110 則為 vlan2 vlan3 為接通著，第28位為1->0001 故 4*27+3=111，
		故還有 vlan111
		*/
		hex_value := hex.EncodeToString(v.Value.([]uint8))
		for i := 0; i < 32*8; i++ {
			if string(hex_value[i]) == "0" {
				continue
			}
			s, err := HexToBin(string(hex_value[i]))
			if err != nil {
				fmt.Printf("HexToBin Error:%v\n", err)
				os.Exit(1)
			}
			for j := 0; j < 4; j++ {
				if string(s[j]) == "1" {
					vlan_name := fmt.Sprintf("vlan%d", i*4+j)
					if strings.EqualFold(native_vlan_table[port_nu], vlan_name) {
						continue
					}
					vlan_table[port_nu] = fmt.Sprintf(vlan_table[port_nu]+" vlan%d", i*4+j)
				}
			}
		}
	}
	return nil
}

/*轉換 ifindex to port number(Cisco & Juniper 通用)*/
func ifDescr(ifindex string) int {
	oids := []string{"1.3.6.1.2.1.2.2.1.2." + ifindex}
	result, err := gosnmp.Default.Get(oids)
	if err != nil {
		log.Fatalf("Get() err: %v", err)
	}
	for _, v := range result.Variables {
		s := strings.Split(string(v.Value.([]byte)), "/") //取得port號
		/*Juniper ge-0/0/1.0
		Cisco  GigabitEthernet0/1*/
		if switch_name == "Juniper" {
			juniper_s := strings.Split(s[len(s)-1], ".") //去掉 .0 (e.g. 14.0 -> 14)
			port_nu, err := strconv.Atoi(juniper_s[0])
			if err != nil {
				fmt.Printf("String to int error:%v\n", err)
				os.Exit(1)
			}
			return port_nu
		} else {
			/*Cisco*/
			if !strings.Contains(s[0], "GigabitEthernet") {
				return -1
			}
			port_nu, err := strconv.Atoi(s[len(s)-1])
			if err != nil {
				fmt.Printf("String to int error:%v\n", err)
				os.Exit(1)
			}
			return port_nu

		}
	}
	return -1
}

/*
dot1dBasePortIfIndex 1.3.6.1.2.1.17.1.4.1.2
Juniper 用 IEEE 802.1D -> IfIndex
*/
func dot1dBasePortIfIndex(dot1d string) string {
	oid := []string{"1.3.6.1.2.1.17.1.4.1.2." + dot1d}
	result, err := gosnmp.Default.Get(oid)
	var ifindex string
	if err != nil {
		log.Fatalf("Get() idot1dBasePortIfIndex err:%v", err)
	}
	for _, v := range result.Variables {
		ifindex = gosnmp.ToBigInt(v.Value).String()
	}
	return ifindex
}

/*
jnxExVlanName 1.3.6.1.4.1.2636.3.40.1.5.1.5.1.2
Juniper 用 將 vlan index -> vlan_name
*/
func jnxExVlanName(index string) string {
	oid := []string{"1.3.6.1.4.1.2636.3.40.1.5.1.5.1.2." + index}
	result, err := gosnmp.Default.Get(oid)
	var vlan_name string
	if err != nil {
		log.Fatalf("Get() jnxExVlanName err:%v", err)
	}
	for _, v := range result.Variables {
		vlan_name = string(v.Value.([]byte))
	}
	return vlan_name
}
func jnxExVlanPortStatus(pdu gosnmp.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".")
	port_nu := ifDescr(oid[len(oid)-1])
	ifindex := dot1dBasePortIfIndex(oid[len(oid)-1])
	port_nu = ifDescr(ifindex)
	vlan_name := jnxExVlanName(oid[len(oid)-2])
	if strings.EqualFold(native_vlan_table[port_nu], vlan_name) {
		return nil
	}
	vlan_table[port_nu] = vlan_table[port_nu] + " " + vlan_name
	return nil
}

/*Port 屬於哪一個 vlan，實作因 OID 不同，又分為 Juniper 與 Cisco*/
func vmVlan(pdu gosnmp.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".") //分割OID
	port_nu := ifDescr(oid[len(oid)-1])
	vlan_table[port_nu] = "vlan" + gosnmp.ToBigInt(pdu.Value).String()
	return nil
}
func set_port_table(pdu gosnmp.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".") //分割OID
	i, err := strconv.Atoi(oid[13])     // OID第14個數字表示 port號 or ifindex
	if err != nil {
		fmt.Printf("String to int error:%v\n", err)
		os.Exit(1)
	}
	/*根據Cisco或是Juniper 儲存 port_table 的 value
	1.3.6.1.2.1.2.2.1.2 ifDescr
	*/
	if switch_name == "Cisco" {
		port_table[i] = port_table[i] + " " + string(pdu.Value.([]byte))
	} else {
		/*Juniper
		查閱 Juniper ifindex
		*/
		i = ifDescr(oid[13]) //第14碼為 Juniper ifindex
		/*要是處理lldpRemSysName(OID 第11碼為8)，將domain_name去掉*/
		if oid[10] == "8" {
			port_table[i] = port_table[i] + "\t" + string(pdu.Value.([]byte))
		} else {
			s := strings.Split(string(pdu.Value.([]byte)), ".")
			port_table[i] = port_table[i] + " " + s[0]
		}
	}
	return nil
}

/*印出交換器內容描述*/
func printswitchDescr(pdu gosnmp.SnmpPDU) error {
	switch pdu.Type {
	case gosnmp.OctetString:
		b := pdu.Value.([]byte)
		s := strings.Split(string(pdu.Value.([]byte)), " ")
		switch_name = s[0] //儲存swtich name
		fmt.Printf("%s\n", string(b))
	default:
		fmt.Printf("TYPE %d: %d\n", pdu.Type, gosnmp.ToBigInt(pdu.Value))
	}
	return nil
}

/*初始化 port_table，此為全域變數，在每次 Ip 更新時刷新一次*/
func intialize_map() {
	port_no = 0
	oids := []string{"1.0.8802.1.1.2.1.3.7.1.3"}
	err := gosnmp.Default.Walk(oids[0], portCount)
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
		os.Exit(1)
	}
	for i := 0; i < 48; i++ {
		port_table[i] = " "
		mode_table[i] = " "
		vlan_table[i] = " "
		native_vlan_table[i] = " "
	}
}

/* 計算 switch port 數目，注意 Cisco 與 Juniper port 開頭分別為 1 與 0*/
func portCount(pdu gosnmp.SnmpPDU) error {
	if switch_name == "Juniper" {
		port_no++
	} else {
		value := strings.Split(string(pdu.Value.([]byte)), "/")
		if strings.Contains(value[0], "Gi") {
			port_no++
		}
	}
	return nil
}
func printTable() {
	fmt.Println("============================================================================================================")
	fmt.Printf("%s%30s%21s%18s%30s\n", "Interface", "RemotePort&Hostname", "PortMode", "NativeVLAN", "VLAN")
	fmt.Println("------------------------------------------------------------------------------------------------------------")
	var i int

	if switch_name == "Cisco" {
		for i = 1; i <= port_no; i++ {
			fmt.Printf("%2d:%36s%21s%18s%30s\n", i, port_table[i], mode_table[i], native_vlan_table[i], vlan_table[i])
		}
	} else {
		for i = 0; i < port_no-1; i++ {
			fmt.Printf("%2d:%36s%21s%18s%30s\n", i, port_table[i], mode_table[i], native_vlan_table[i], vlan_table[i])

		}

	}
}
func HexToBin(hex string) (string, error) {
	ui, err := strconv.ParseUint(hex, 16, 64)
	if err != nil {
		return "", err
	}

	// %04b indicates base 2, zero padded, with 4 characters
	return fmt.Sprintf("%04b", ui), nil
}
