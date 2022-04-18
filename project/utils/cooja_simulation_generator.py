import ipaddress
import copy
from xml.dom.minidom import parseString
from sys import argv

template_csc_file = """\
<?xml version="1.0" encoding="UTF-8"?>
<simconf>
  <project EXPORT="discard">[APPS_DIR]/mrm</project>
  <project EXPORT="discard">[APPS_DIR]/mspsim</project>
  <project EXPORT="discard">[APPS_DIR]/avrora</project>
  <project EXPORT="discard">[APPS_DIR]/serial_socket</project>
  <project EXPORT="discard">[APPS_DIR]/powertracker</project>
  <simulation>
    <title>UDP Simulation</title>
    <randomseed>generated</randomseed>
    <motedelay_us>1000000</motedelay_us>
    <radiomedium>
      org.contikios.cooja.radiomediums.UDGM
      <transmitting_range>50.0</transmitting_range>
      <interference_range>99.0</interference_range>
      <success_ratio_tx>1.0</success_ratio_tx>
      <success_ratio_rx>1.0</success_ratio_rx>
    </radiomedium>
    <events>
      <logoutput>40000</logoutput>
    </events>
    <motetype>
      org.contikios.cooja.contikimote.ContikiMoteType
      <identifier>mtype522</identifier>
      <description>Proxy</description>
      <source>[CONTIKI_DIR]/project/proxy/proxy.c</source>
      <commands>make proxy.cooja TARGET=cooja -j8</commands>
      <moteinterface>org.contikios.cooja.interfaces.Position</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.Battery</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiVib</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiMoteID</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiRS232</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiBeeper</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.RimeAddress</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiIPAddress</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiRadio</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiButton</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiPIR</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiClock</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiLED</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiCFS</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiEEPROM</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.Mote2MoteRelations</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.MoteAttributes</moteinterface>
      <symbols>false</symbols>
    </motetype>
    <motetype>
      org.contikios.cooja.contikimote.ContikiMoteType
      <identifier>mtype198</identifier>
      <description>Client</description>
      <source>[CONTIKI_DIR]/project/client/udp-client.c</source>
      <commands>make udp-client.cooja TARGET=cooja -j8</commands>
      <moteinterface>org.contikios.cooja.interfaces.Position</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.Battery</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiVib</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiMoteID</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiRS232</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiBeeper</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.RimeAddress</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiIPAddress</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiRadio</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiButton</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiPIR</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiClock</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiLED</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiCFS</moteinterface>
      <moteinterface>org.contikios.cooja.contikimote.interfaces.ContikiEEPROM</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.Mote2MoteRelations</moteinterface>
      <moteinterface>org.contikios.cooja.interfaces.MoteAttributes</moteinterface>
      <symbols>false</symbols>
    </motetype>
  </simulation>
  <plugin>
    org.contikios.cooja.plugins.SimControl
    <width>280</width>
    <z>1</z>
    <height>160</height>
    <location_x>400</location_x>
    <location_y>0</location_y>
  </plugin>
  <plugin>
    org.contikios.cooja.plugins.Visualizer
    <plugin_config>
      <moterelations>true</moterelations>
      <skin>org.contikios.cooja.plugins.skins.IDVisualizerSkin</skin>
      <skin>org.contikios.cooja.plugins.skins.GridVisualizerSkin</skin>
      <skin>org.contikios.cooja.plugins.skins.TrafficVisualizerSkin</skin>
      <skin>org.contikios.cooja.plugins.skins.UDGMVisualizerSkin</skin>
      <viewport>1.155021851567949 0.0 0.0 1.155021851567949 139.20443484841425 26.053056915508364</viewport>
    </plugin_config>
    <width>400</width>
    <z>2</z>
    <height>205</height>
    <location_x>1</location_x>
    <location_y>1</location_y>
  </plugin>
  <plugin>
    org.contikios.cooja.plugins.LogListener
    <plugin_config>
      <filter />
      <formatted_time />
      <coloring />
    </plugin_config>
    <width>1318</width>
    <z>4</z>
    <height>796</height>
    <location_x>-1</location_x>
    <location_y>207</location_y>
  </plugin>
  <plugin>
    org.contikios.cooja.serialsocket.SerialSocketServer
    <mote_arg>0</mote_arg>
    <plugin_config>
      <port>60001</port>
      <bound>true</bound>
    </plugin_config>
    <width>362</width>
    <z>3</z>
    <height>116</height>
    <location_x>680</location_x>
    <location_y>0</location_y>
  </plugin>
  <plugin>
    org.contikios.cooja.plugins.ScriptRunner
    <plugin_config>
      <scriptfile>[CONFIG_DIR]/script.js</scriptfile>
      <active>true</active>
    </plugin_config>
    <width>600</width>
    <z>0</z>
    <height>700</height>
    <location_x>723</location_x>
    <location_y>56</location_y>
  </plugin>
</simconf>
"""

moteXML = """\
<mote>
  <interface_config>
    org.contikios.cooja.interfaces.Position
    <x>0.0</x>
    <y>0.0</y>
    <z>0.0</z>
  </interface_config>
  <interface_config>
    org.contikios.cooja.contikimote.interfaces.ContikiMoteID
    <id>1</id>
  </interface_config>
  <interface_config>
    org.contikios.cooja.contikimote.interfaces.ContikiRadio
    <bitrate>250.0</bitrate>
  </interface_config>
  <interface_config>
    org.contikios.cooja.contikimote.interfaces.ContikiEEPROM
    <eeprom>AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==</eeprom>
  </interface_config>
  <motetype_identifier>mtype</motetype_identifier>
</mote>
"""

def add_mote(simulation, motetype_identifier, id, x, y):
    # should use xpath, but it works well 🙃
    mote = copy.deepcopy(parseString(moteXML))
    mote.getElementsByTagName("motetype_identifier")[0].firstChild.data = motetype_identifier
    interfaces_config = mote.getElementsByTagName("interface_config")
    interfaces_config[0].getElementsByTagName("x")[0].firstChild.data = x
    interfaces_config[0].getElementsByTagName("y")[0].firstChild.data = y
    interfaces_config[1].getElementsByTagName("id")[0].firstChild.data = id
    simulation.getElementsByTagName("simulation")[0].appendChild(copy.deepcopy(mote.firstChild))

def find_motetype_by_description(simulation, description):
    motetypes = simulation.getElementsByTagName("motetype")
    for motetype in motetypes:
        try:
            if motetype.getElementsByTagName("description")[0].firstChild.data == description:
                return motetype
        except:
            pass
    raise ValueError(f"Could not find motetype with description: {description}")
    
def get_motetype_identifier(motetype):
    return motetype.getElementsByTagName("identifier")[0].firstChild.data

def set_mote_commands(motetype, server_addr, port: int):
    commands = motetype.getElementsByTagName("commands")[0].firstChild
    commands.data += f" UDP_SERVER_PORT={port}"
    addr_macro = "{{"
    for i in range(len(server_addr.packed) - 1):
        addr_macro += f"{server_addr.packed[i]},"
    addr_macro += str(server_addr.packed[-1])
    addr_macro += "}}"

    commands.data += f" UDP_SERVER_ADDR={addr_macro}"

def find_transmission_range(simulation):
    return float(simulation.getElementsByTagName("transmitting_range")[0].firstChild.data)

def set_port(simulation, port):
    simulation.getElementsByTagName("port")[0].firstChild.data = port

def change_script_filename(simulation, script_filename):
    scriptNode = simulation.getElementsByTagName("scriptfile")[0]
    scriptNode.data = f"[CONFIG_DIR]/{script_filename}"

def linear_topology_simulation(motes_count: int, first_mote_id: int, server_addr: str, server_port: int, tunslip_port: int) -> str:
    simulation = parseString(template_csc_file)
    client = find_motetype_by_description(simulation, "Client")
    proxy = find_motetype_by_description(simulation, "Proxy")
    client_identifier = get_motetype_identifier(client)
    proxy_identifier = get_motetype_identifier(proxy)
    transmission_range = find_transmission_range(simulation)

    add_mote(simulation, proxy_identifier, first_mote_id, 0.0, 0.0)
    for i in range(first_mote_id + 1, first_mote_id + motes_count):
        add_mote(simulation, client_identifier, i, transmission_range * (i - first_mote_id), 0.0)
    set_port(simulation, tunslip_port)
    server_addr = ipaddress.ip_address(server_addr)
    set_mote_commands(client, server_addr, server_port)
    set_mote_commands(proxy, server_addr, server_port)
    return simulation.toxml()

def grid_topology_simulation(motes_count: int, first_mote_id: int, server_addr: str, server_port: int, tunslip_port: int) -> str:
    simulation = parseString(template_csc_file)
    client = find_motetype_by_description(simulation, "Client")
    proxy = find_motetype_by_description(simulation, "Proxy")
    client_identifier = get_motetype_identifier(client)
    proxy_identifier = get_motetype_identifier(proxy)
    transmission_range = find_transmission_range(simulation)

    add_mote(simulation, proxy_identifier, first_mote_id, 0.0, 0.0)
    for i in range(first_mote_id + 1, first_mote_id + motes_count):
        x = ((i-1) % 3) * transmission_range
        y = ((i-1) // 3) * transmission_range
        add_mote(simulation, client_identifier, i, x, y)
    set_port(simulation, tunslip_port)
    server_addr = ipaddress.ip_address(server_addr)
    set_mote_commands(client, server_addr, server_port)
    set_mote_commands(proxy, server_addr, server_port)
    return simulation.toxml()


if __name__ == "__main__":
    print(linear_topology_simulation(int(argv[1]), int(argv[2]), "fd00::1", 3000, 60001))
