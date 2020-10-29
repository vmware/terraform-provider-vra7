package vra7

var (
	validAuthResponse = `{
		"expires":"2019-02-26T03:32:35.000Z",
		"id":"MTU1MTEyMzE1NTc5ODpiYTZkYjdhNjZlNGNkYjZmZTBiMjp0ZW5hbnQ6cWV1c2VybmFtZTpmcml0ekBjb2tlLnNxYS1ob3Jpem9uLmxvY2FsZXhwaXJhdGlvbjoxNTUxMTUxOTU1MDAwOmMyNGVjNTFiNzE1OTJhZDZjNTljMTUwMDkxMjcyNzUyZDkzNzQ0ODRkMTVlZGFhNWM0MDhjYmQ3YTM2MTljZGNiNjM3MjM1NmY1MzZlYTk1YzUyMGZiZDVjMTkzMzg3YjQzZmMwNmNlMGI5YjJkZmIwNzhlZGU2NzdiNTk3MWFk",
		"tenant":"qe"
	 }`

	errorAuthResponse = `{
		"errors":[
		   {  
			  "code":90135,
			  "source":null,
			  "message":"Unable to authenticate user fritz@coke.sqa-horizon.local in tenant q.",
			  "systemMessage":"90135-Unable to authenticate user fritz@coke.sqa-horizon.local in tenant q.",
			  "moreInfoUrl":null
		   }
		]
	 }`

	mockRequestTemplate = `{
		"type":"com.vmware.vcac.catalog.domain.request.CatalogItemProvisioningRequest",
		"catalogItemId":"dhbh-jhdv-ghdv-dhvdd",
		"requestedFor":"mock@mock.com",
		"businessGroupId":"vcdo-vdgvcd-hgvdc",
		"description":"this is a mock template",
		"reasons":"for unit test",
		"data":{
		   "_leaseDays":null,
		   "_number_of_instances":2,
		   "mock.test.machine1":{
			  "classId":"Blueprint.Component.Declaration",
			  "componentId":null,
			  "componentTypeId":"com.vmware.csp.component.cafe.composition",
			  "data":{
				 "_cluster":1,
				 "_hasChildren":false,
				 "cpu":1,
				 "datacenter_location":null,
				 "description":null,
				 "disks":[
					{
					   "classId":"Infrastructure.Compute.Machine.MachineDisk",
					   "componentId":null,
					   "componentTypeId":"com.vmware.csp.iaas.blueprint.service",
					   "data":{
						  "capacity":8,
						  "id":68987664545,
						  "initial_location":"",
						  "is_clone":false,
						  "label":"",
						  "storage_reservation_policy":"",
						  "userCreated":true,
						  "volumeId":0
					   },
					   "typeFilter":null
					}
				 ],
				 "display_location":false,
				 "guest_customization_specification":null,
				 "machine_prefix":null,
				 "max_network_adapters":-1,
				 "max_per_user":0,
				 "max_volumes":60,
				 "memory":4096,
				 "nics":null,
				 "os_arch":"x86_64",
				 "os_distribution":null,
				 "os_type":"Linux",
				 "os_version":null,
				 "ovfAuthNeeded":false,
				 "ovf_proxy_endpoint":null,
				 "ovf_url":null,
				 "ovf_url_pwd":null,
				 "ovf_url_username":null,
				 "ovf_use_proxy":false,
				 "property_groups":null,
				 "reservation_policy":null,
				 "security_groups":[
	 
				 ],
				 "security_tags":[
	 
				 ],
				 "snapshot_name":null,
				 "source_machine":null,
				 "source_machine_external_snapshot":null,
				 "source_machine_name":null,
				 "source_machine_vmsnapshot":null,
				 "storage":8
			  },
			  "typeFilter":"mock_type*mock.machine1"
		   },
		   "machine2":{
			  "classId":"Blueprint.Component.Declaration",
			  "componentId":null,
			  "componentTypeId":"com.vmware.csp.component.cafe.composition",
			  "data":{
				 "_cluster":1,
				 "_hasChildren":false,
				 "cpu":1,
				 "datacenter_location":null,
				 "description":null,
				 "disks":[
					{
					   "classId":"Infrastructure.Compute.Machine.MachineDisk",
					   "componentId":null,
					   "componentTypeId":"com.vmware.csp.iaas.blueprint.service",
					   "data":{
						  "capacity":8,
						  "custom_properties":null,
						  "id":4487629629,
						  "initial_location":"",
						  "is_clone":true,
						  "label":"Hard disk 1",
						  "storage_reservation_policy":"",
						  "userCreated":false,
						  "volumeId":0
					   },
					   "typeFilter":null
					}
				 ],
				 "display_location":false,
				 "guest_customization_specification":null,
				 "location.loc":"",
				 "machine_prefix":null,
				 "max_network_adapters":-1,
				 "max_per_user":0,
				 "max_volumes":60,
				 "memory":1024,
				 "nics":null,
				 "os_arch":"x86_64",
				 "os_distribution":null,
				 "os_type":"Linux",
				 "os_version":null,
				 "ovfAuthNeeded":false,
				 "ovf_proxy_endpoint":null,
				 "ovf_url":null,
				 "ovf_url_pwd":null,
				 "ovf_url_username":null,
				 "ovf_use_proxy":false,
				 "property_groups":null,
				 "reservation_policy":null,
				 "security_groups":[
	 
				 ],
				 "security_tags":[
	 
				 ],
				 "snapshot_name":null,
				 "source_machine_external_snapshot":null,
				 "source_machine_vmsnapshot":null,
				 "storage":8
			  },
			  "typeFilter":"mock_type@machine2"
		   }
		}
	 }`
)
