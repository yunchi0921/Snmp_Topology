package main

import (
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
	port_no     int
	switch_name string
	port_table  map[int]string
)

func main() {
	Ip := []string{"192.168.1.1", "192.168.1.3"}
	port_table = make(map[int]string)

	/*一次進行一個機器*/
	for i := 0; i < len(Ip); i++ {
		/*連線*/
		gosnmp.Default.Target = Ip[i]
		err := gosnmp.Default.Connect()
		if err != nil {
			log.Fatalf("Connect() err:%v", err)
		}
		defer gosnmp.Default.Conn.Close()

		/*初始化 port_table，在每一個新的機器(Ip)都要重製表。*/
		intialize_map()

		/*要查詢的OID
		1.0.8802.1.1.2.1.3.4		lldpLocSysDesc
		1.0.8802.1.1.2.1.4.1.1.8	lldpRemPortDesc
		1.0.8802.1.1.2.1.4.1.1.9	ldpRemSysName
		*/
		oids := []string{"1.0.8802.1.1.2.1.3.4", "1.0.8802.1.1.2.1.4.1.1.8", "1.0.8802.1.1.2.1.4.1.1.9"}
		for i := 0; i < len(oids); i++ {
			switch i {
			case 0:
				err = gosnmp.Default.Walk(oids[i], printValue)
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
		/*port vlan 查詢，因為 Cisco & Juniper OID 不同，對於 Port 使用的 index 方法也不同，
		必須分開時做，用前面儲存的 switch_name 進行判別式，區分出兩種機器實作
		1.3.6.1.4.1.9.9.68.1.2.2.1.2		CISCO-VLAN-MEMBERSHIP-MIB
		1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3	jnxExVlanPortStatus*/
		oids = []string{"1.3.6.1.4.1.9.9.68.1.2.2.1.2", "1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3"}
		if switch_name == "Cisco" {
			err = gosnmp.Default.Walk(oids[0], vlan_set_table)
			if err != nil {
				fmt.Printf("vlan_set_table Walk Error:%v\n", err)
				os.Exit(1)
			}
		} else {

		}

		printTable()

	}
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
			port_no, err := strconv.Atoi(juniper_s[0])
			if err != nil {
				fmt.Printf("String to int error:%v\n", err)
				os.Exit(1)
			}
			return port_no
		} else {
			/*Cisco*/
			port_no, err := strconv.Atoi(s[len(s)-1])
			if err != nil {
				fmt.Printf("String to int error:%v\n", err)
				os.Exit(1)
			}
			return port_no
		}
	}
	return -1
}
func vlan_set_table(pdu gosnmp.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".") //分割OID
	port_no := ifDescr(oid[len(oid)-1])
	port_table[port_no] = port_table[port_no] + " vlan" + gosnmp.ToBigInt(pdu.Value).String()
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
		/*查閱 Juniper ifindex*/
		oids := []string{"1.3.6.1.2.1.2.2.1.2." + oid[13]}
		result, err := gosnmp.Default.Get(oids)
		if err != nil {
			log.Fatalf("Get() err: %v", err)
		}
		for _, v := range result.Variables {
			s := strings.Split(string(v.Value.([]byte)), "/") //取得port號
			s = strings.Split(s[len(s)-1], ".")               //去掉 .0 (e.g. 14.0 -> 14)
			i, err := strconv.Atoi(s[0])
			if err != nil {
				fmt.Printf("String to int error:%v\n", err)
				os.Exit(1)
			}
			/*要是處理lldpRemSysName(OID 第11碼為8)，將domain_name去掉*/
			if oid[10] == "8" {
				port_table[i] = port_table[i] + " " + string(pdu.Value.([]byte))
			} else {
				s = strings.Split(string(pdu.Value.([]byte)), ".")
				port_table[i] = port_table[i] + " " + s[0]
			}

		}
	}
	return nil
}
func printValue(pdu gosnmp.SnmpPDU) error {
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
func intialize_map() {
	port_no = 0
	oids := []string{"1.0.8802.1.1.2.1.3.7.1.3"}
	err := gosnmp.Default.Walk(oids[0], portCount)
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
		os.Exit(1)
	}
	for i := 0; i < 48; i++ {
		port_table[i] = ""
	}
}
func portCount(pdu gosnmp.SnmpPDU) error {
	port_no++
	return nil
}
func printTable() {
	var i int
	if switch_name == "Cisco" {
		for i = 1; i <= port_no/2; i++ {
			fmt.Printf("%2d:%30s\t%2d:%30s\n", i, port_table[i], i+port_no/2, port_table[i+port_no/2])
		}
	} else {
		for i = 0; i < port_no/2; i++ {
			fmt.Printf("%2d:%30s\t%2d:%30s\n", i, port_table[i], i+port_no/2, port_table[i+port_no/2])
		}

	}
}
