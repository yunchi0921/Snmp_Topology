# Switch Topology

Switch topology is a program in GO. It gives u information of all of switch connected each other.It provides Hostname, Remote Port Index, PortMode ,VLAN and Native VLan. It is implemented by [soniah/gosnmp](https://github.com/soniah/gosnmp)

## Usage

Here is `snmp.go` . You should change your IP in the IP array in the beginning of main function.
e.g.
```clike
	Ip := []string{"Your switch IP address1", "Your switch IP address2"}  //change your IP that you want to monitor
	port_table = make(map[int]string)
	mode_table = make(map[int]string)
	vlan_table = make(map[int]string)
```

`go run snmp.go`

Running the default IP gives the following output(with my switch setting one Cisco Catalyst2960 and Juniper ex2200)

```
Juniper Networks, Inc. ex2200-48t-4g , version 12.3R1.7 Build date: 2013-01-26 01:45:11 UTC
============================================================================================================
Interface           RemotePort&Hostname             PortMode        NativeVLAN                          VLAN
------------------------------------------------------------------------------------------------------------
 0:         GigabitEthernet0/1 switch-3               access                                           vlan1
 1:                                                   access                                           vlan1
 2:         GigabitEthernet0/2 switch-3                trunk          testvlan                         vlan1
 3:                                                   access                                           vlan1
 4:                                                   access                                           vlan1
 5:                                                   access                                           vlan1
 6:                                                   access                                           vlan1
 7:                                                   access                                           vlan1
 8:                                                   access                                           vlan1
 9:                                                   access                                           vlan1
10:                                                   access                                           vlan1
11:                                                   access                                           vlan1
12:                                                   access                                           vlan1
13:                                                   access                                           vlan1
14:                                                   access                                           vlan1
15:                                                   access                                           vlan1
16:                                                   access                                           vlan1
17:                                                   access                                           vlan1
18:                                                   access                                           vlan1
19:                                                   access                                           vlan1
20:        GigabitEthernet0/11 switch-3               access                                           vlan1
21:                                                   access                                           vlan1
22:                                                   access                                           vlan1
23:                                                   access                                           vlan1
24:                                                   access                                           vlan1
25:                                                   access                                           vlan1
26:                                                   access                                           vlan1
27:                                                   access                                           vlan1
28:                                                   access                                           vlan1
29:                                                   access                                           vlan1
30:                                                   access                                           vlan1
31:                                                   access                                           vlan1
32:                                                   access                                           vlan1
33:                                                   access                                           vlan1
34:                                                   access                                           vlan1
35:                                                   access                                           vlan1
36:                                                   access                                           vlan1
37:                                                   access                                           vlan1
38:                                                   access                                           vlan1
39:                                                   access                                           vlan1
40:                                                   access                                           vlan1
41:                                                   access                                           vlan1
42:                                                   access                                           vlan1
43:                                                   access                                           vlan1
44:                                                   access                                           vlan1
45:                                                   access                                           vlan1
46:                                                   access                                           vlan1
47:                                                   access                                           vlan1
Cisco IOS Software, C2960 Software (C2960-LANBASEK9-M), Version 12.2(53)SE2, RELEASE SOFTWARE (fc3)
Technical Support: http://www.cisco.com/techsupport
Copyright (c) 1986-2010 by Cisco Systems, Inc.
Compiled Wed 21-Apr-10 05:52 by prod_rel_team
============================================================================================================
Interface           RemotePort&Hostname             PortMode        NativeVLAN                          VLAN
------------------------------------------------------------------------------------------------------------
 1:                 ge-0/0/0.0 switch-1               access                                           vlan1
 2:                 ge-0/0/2.0 switch-1                trunk           vlan111                   vlan1 vlan2
 3:                                                   access                                           vlan1
 4:                                                   access                                           vlan1
 5:               en5 macbook-pro.local               access                                           vlan2
 6:                                                   access                                           vlan1
 7:                                                   access                                           vlan1
 8:                                                   access                                           vlan1
 9:                                                   access                                           vlan1
10:                                                   access                                           vlan1
11:                ge-0/0/20.0 switch-1               access                                           vlan1
12:                                                   access                                           vlan1
13:                                                   access                                           vlan1
14:                                                   access                                           vlan1
15:                                                   access                                           vlan1
16:                                                   access                                           vlan1
17:                                                   access                                           vlan1
18:                                                   access                                           vlan1
19:                                                   access                                           vlan2
20:                                                   access                                           vlan1
21:                                                   access                                           vlan1
22:                                                   access                                           vlan1
23:                                                   access                                           vlan1
24:                                                   access                                           vlan1
```
