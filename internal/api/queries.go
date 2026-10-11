package api

// GraphQL queries for Unraid API
// Based on actual schema from https://docs.unraid.net/API/

const InfoQuery = `
query {
	info {
		versions { core { unraid api kernel } }
		os {
			platform
			distro
			release
			uptime
			hostname
		}
		cpu {
			manufacturer
			brand
			cores
			speed
		}
	}
}
`

const ArrayStatusQuery = `
query {
 array {
  state
  capacity { kilobytes { free used total } }
  disks { id name device size status temp type numErrors }
  parities { id name device size status temp type numErrors }
  caches { id name device size status temp type numErrors }
  boot { id name device size status temp type numErrors }
  parityCheckStatus { date duration status errors progress running paused correcting }
 }
}
`

const ArraySetStateMutation = `
mutation($input: ArrayStateInput!) {
	array {
		setState(input: $input) {
			state
			capacity {
				kilobytes {
					free
					used
					total
				}
			}
			disks {
				id
				name
				device
				size
				status
				temp
				type
			}
		}
	}
}
`

const ArrayAddDiskMutation = `
mutation($input: ArrayDiskInput!) {
	array {
		addDiskToArray(input: $input) {
			state
			disks {
				id
				name
				device
				size
				status
				temp
				type
			}
		}
	}
}
`

const ArrayMountDiskMutation = `
mutation($id: PrefixedID!) {
	array {
		mountArrayDisk(id: $id) {
			id
			name
			device
			size
			status
			temp
			type
		}
	}
}
`

const ArrayUnmountDiskMutation = `
mutation($id: PrefixedID!) {
	array {
		unmountArrayDisk(id: $id) {
			id
			name
			device
			size
			status
			temp
			type
		}
	}
}
`

const ArrayClearDiskStatsMutation = `
mutation($id: PrefixedID!) {
	array {
		clearArrayDiskStatistics(id: $id)
	}
}
`

const DockerContainersQuery = `
query {
	docker {
		containers {
			id
			names
			state
			status
			image
			created
			lanIpPorts
			autoStart
			autoStartOrder
			autoStartWait
			ports {
				privatePort
				publicPort
				type
			}
			hostConfig {
				networkMode
			}
			isOrphaned
			isUpdateAvailable
			isRebuildReady
			projectUrl
			registryUrl
			supportUrl
			iconUrl
			webUiUrl
			shell
			tailscaleEnabled
		}
	}
}
`

const DockerContainerQuery = `
query($id: PrefixedID!) {
	docker {
		container(id: $id) {
			id
			names
			state
			status
			image
			imageId
			command
			created
			lanIpPorts
			sizeRootFs
			sizeRw
			sizeLog
			autoStart
			autoStartOrder
			autoStartWait
			ports {
				privatePort
				publicPort
				type
			}
			hostConfig {
				networkMode
			}
			networkSettings
			mounts
			isOrphaned
			isUpdateAvailable
			isRebuildReady
			templatePath
			projectUrl
			registryUrl
			supportUrl
			iconUrl
			webUiUrl
			shell
			templatePorts {
				privatePort
				publicPort
				type
			}
			tailscaleEnabled
		}
	}
}
`

const DockerLogsQuery = `
query($id: PrefixedID!, $since: DateTime, $tail: Int) {
	docker {
		logs(id: $id, since: $since, tail: $tail) {
			containerId
			cursor
			lines {
				timestamp
				message
			}
		}
	}
}
`

const DockerStartMutation = `
mutation($id: PrefixedID!) {
	docker {
		start(id: $id) {
			id
			names
			state
			status
			image
			autoStart
		}
	}
}
`

const DockerStopMutation = `
mutation($id: PrefixedID!) {
	docker {
		stop(id: $id) {
			id
			names
			state
			status
			image
			autoStart
		}
	}
}
`

const DockerRestartMutation = `
mutation($id: PrefixedID!) {
	docker {
		restart(id: $id) {
			id
			names
			state
			status
			image
			autoStart
		}
	}
}
`

const DockerPauseMutation = `
mutation($id: PrefixedID!) {
	docker {
		pause(id: $id) {
			id
			names
			state
			status
			image
			autoStart
		}
	}
}
`

const NetworkMetricsQuery = `
query {
	metrics {
		network {
			id
			name
			operstate
			bytesReceived
			bytesSent
			packetsReceived
			packetsSent
			receiveErrors
			transmitErrors
			receiveDropped
			transmitDropped
			rxSec
			txSec
			utilizationPercent
			lastUpdated
		}
	}
}
`

const DockerUnpauseMutation = `
mutation($id: PrefixedID!) {
	docker {
		unpause(id: $id) {
			id
			names
			state
			status
			image
			autoStart
		}
	}
}
`

const DockerUpdateMutation = `
mutation($id: PrefixedID!) {
	docker {
		updateContainer(id: $id) {
			id
			names
			state
			status
			image
			isUpdateAvailable
			isRebuildReady
		}
	}
}
`

const DockerUpdateAllMutation = `
mutation {
	docker {
		updateAllContainers {
			id
			names
			state
			status
			image
			isUpdateAvailable
			isRebuildReady
		}
	}
}
`

const DockerRemoveMutation = `
mutation($id: PrefixedID!, $withImage: Boolean) {
	docker {
		removeContainer(id: $id, withImage: $withImage)
	}
}
`

const DockerAutostartMutation = `
mutation($entries: [DockerAutostartEntryInput!]!, $persist: Boolean) {
	docker {
		updateAutostartConfiguration(entries: $entries, persistUserPreferences: $persist)
	}
}
`

const APIKeysQuery = `
query {
	apiKeys {
		id
		name
		description
		roles
		createdAt
		permissions {
			resource
			actions
		}
	}
}
`

const APIKeyMetadataQuery = `
query {
	apiKeyPossibleRoles
	apiKeyPossiblePermissions {
		resource
		actions
	}
	getAvailableAuthActions
}
`

const CreateAPIKeyMutation = `
mutation($input: CreateApiKeyInput!) {
	apiKey {
		create(input: $input) {
			id
			key
			name
			description
			roles
			createdAt
			permissions {
				resource
				actions
			}
		}
	}
}
`

const UpdateAPIKeyMutation = `
mutation($input: UpdateApiKeyInput!) {
	apiKey {
		update(input: $input) {
			id
			name
			description
			roles
			createdAt
			permissions {
				resource
				actions
			}
		}
	}
}
`

const DeleteAPIKeyMutation = `
mutation($input: DeleteApiKeyInput!) {
	apiKey {
		delete(input: $input)
	}
}
`

const AddAPIKeyRoleMutation = `
mutation($input: AddRoleForApiKeyInput!) {
	apiKey {
		addRole(input: $input)
	}
}
`

const RemoveAPIKeyRoleMutation = `
mutation($input: RemoveRoleFromApiKeyInput!) {
	apiKey {
		removeRole(input: $input)
	}
}
`

const SharesQuery = `
query {
	shares {
		name
		comment
		free
		used
	}
}
`

const NotificationsQuery = `
query($filter: NotificationFilter!) {
 notifications {
  overview { unread { total info warning alert } archive { total info warning alert } }
  list(filter: $filter) { id subject importance timestamp type }
 }
}
`

const NotificationAlertsQuery = `
query {
 notifications {
  overview { unread { total info warning alert } archive { total info warning alert } }
  warningsAndAlerts { id subject importance timestamp type }
 }
}
`

const LogFilesQuery = `
query {
	logFiles {
		name
		path
		size
		modifiedAt
	}
}
`

const LogFileQuery = `
query($path: String!, $lines: Int, $startLine: Int) {
	logFile(path: $path, lines: $lines, startLine: $startLine) {
		path
		content
		totalLines
		startLine
	}
}
`

const SettingsQuery = `
query {
	isSSOEnabled
	settings {
		id
		api {
			version
			extraOrigins
			sandbox
			ssoSubIds
			plugins
		}
		sso {
			id
			oidcProviders {
				id
				name
				clientId
				issuer
				authorizationEndpoint
				tokenEndpoint
				jwksUri
				scopes
				buttonText
				buttonVariant
			}
		}
		unified {
			values
		}
	}
}
`

const UpdateSettingsMutation = `
mutation($input: JSON!) {
	updateSettings(input: $input) {
		restartRequired
		warnings
		values
	}
}
`

const OIDCProvidersQuery = `
query {
	isSSOEnabled
	oidcProviders {
		id
		name
		clientId
		issuer
		authorizationEndpoint
		tokenEndpoint
		jwksUri
		scopes
		buttonText
		buttonVariant
	}
	publicOidcProviders {
		id
		name
		buttonText
		buttonIcon
		buttonVariant
		buttonStyle
	}
}
`

const OIDCConfigurationQuery = `
query {
	oidcConfiguration {
		defaultAllowedOrigins
		providers {
			id
			name
			clientId
			issuer
			authorizationEndpoint
			tokenEndpoint
			jwksUri
			scopes
			buttonText
			buttonVariant
		}
	}
}
`

const ValidateOIDCSessionQuery = `
query($token: String!) {
	validateOidcSession(token: $token) {
		valid
		username
	}
}
`

const UPSQuery = `
query {
 upsDevices {
  name model status
  battery { chargeLevel estimatedRuntime }
  power { loadPercentage currentPower }
 }
}
`

const DisksQuery = `
query {
 disks { id device name type size smartStatus temperature }
}
`

const ParityStatusQuery = `
query {
 array { parityCheckStatus { date duration status errors progress running paused correcting } }
}
`

const ParityHistoryQuery = `
query {
 parityHistory { date duration status errors progress running paused correcting }
}
`

const CPUMetricsQuery = `
query { metrics { cpu { percentTotal } } }
`

const MemoryMetricsQuery = `
query { metrics { memory { total available active percentTotal } } }
`

const TemperatureMetricsQuery = `
query { metrics { temperature { sensors { name type current { value unit status } } } } }
`

const SystemMetricsQuery = `
query {
 metrics {
  cpu { percentTotal }
  memory { total available active percentTotal }
  temperature { sensors { name type current { value unit status } } }
 }
}
`

const ContainerHealthQuery = `
query { docker { containers { id names state status image autoStart } } }
`

const VMListQuery = `
query { vms { id domains { id name state } } }
`

const VMStartMutation = `mutation($id: PrefixedID!) { vm { start(id: $id) } }`
const VMStopMutation = `mutation($id: PrefixedID!) { vm { stop(id: $id) } }`
const VMPauseMutation = `mutation($id: PrefixedID!) { vm { pause(id: $id) } }`
const VMResumeMutation = `mutation($id: PrefixedID!) { vm { resume(id: $id) } }`
const VMForceStopMutation = `mutation($id: PrefixedID!) { vm { forceStop(id: $id) } }`
const VMRebootMutation = `mutation($id: PrefixedID!) { vm { reboot(id: $id) } }`
const VMResetMutation = `mutation($id: PrefixedID!) { vm { reset(id: $id) } }`

const ParityStartMutation = `mutation($correct: Boolean!) { parityCheck { start(correct: $correct) } }`
const ParityPauseMutation = `mutation { parityCheck { pause } }`
const ParityResumeMutation = `mutation { parityCheck { resume } }`
const ParityCancelMutation = `mutation { parityCheck { cancel } }`

const NotificationArchiveManyMutation = `mutation($ids: [PrefixedID!]!) { archiveNotifications(ids: $ids) { unread { total info warning alert } archive { total info warning alert } } }`
const NotificationArchiveAllMutation = `mutation($importance: NotificationImportance) { archiveAll(importance: $importance) { unread { total info warning alert } archive { total info warning alert } } }`
const NotificationUnarchiveManyMutation = `mutation($ids: [PrefixedID!]!) { unarchiveNotifications(ids: $ids) { unread { total info warning alert } archive { total info warning alert } } }`
const NotificationUnarchiveAllMutation = `mutation($importance: NotificationImportance) { unarchiveAll(importance: $importance) { unread { total info warning alert } archive { total info warning alert } } }`
const NotificationDeleteMutation = `mutation($id: PrefixedID!, $type: NotificationType!) { deleteNotification(id: $id, type: $type) { unread { total info warning alert } archive { total info warning alert } } }`
const NotificationDeleteArchivedMutation = `mutation { deleteArchivedNotifications { unread { total info warning alert } archive { total info warning alert } } }`
const NotificationCreateMutation = `mutation($input: NotificationData!) { createNotification(input: $input) { id subject importance timestamp type } }`
const NotificationNotifyIfUniqueMutation = `mutation($input: NotificationData!) { notifyIfUnique(input: $input) { id subject importance timestamp type } }`
const NotificationRecalculateMutation = `mutation { recalculateOverview { unread { total info warning alert } archive { total info warning alert } } }`

const DockerUpdateManyMutation = `
mutation($ids: [PrefixedID!]!) {
 docker { updateContainers(ids: $ids) { id names state status image autoStart isUpdateAvailable isRebuildReady } }
}
`

const DockerRefreshDigestsMutation = `mutation { refreshDockerDigests }`

const FlashBackupMutation = `
mutation($input: InitiateFlashBackupInput!) { initiateFlashBackup(input: $input) { status jobId } }
`

const ServicesQuery = `
query { services { id name online version uptime { timestamp } } }
`

const ServerQuery = `
query { server { id name comment status lanip wanip localurl remoteurl } }
`

const NetworkInterfacesQuery = `
query {
 networkInterfaces {
  id name description macAddress mtu speed duplex virtual operstate type vlanId
  ipAddress netmask gateway useDhcp ipv6Address ipv6Netmask ipv6Gateway useDhcp6
 }
}
`

const PluginsQuery = `
query { plugins { name version hasApiModule hasCliModule } installedUnraidPlugins }
`

const SystemTimeQuery = `
query { systemTime { currentTime timeZone useNtp ntpServers } }
`

const TimeZoneOptionsQuery = `
query { timeZoneOptions { value label } }
`

const UpdateSystemTimeMutation = `
mutation($input: UpdateSystemTimeInput!) { updateSystemTime(input: $input) { currentTime timeZone useNtp ntpServers } }
`

const UPSConfigurationQuery = `
query {
 upsConfiguration {
  service upsCable customUpsCable upsType device overrideUpsCapacity batteryLevel minutes timeout killUps
  nisIp netServer upsName modelName
 }
}
`

const ConfigureUPSMutation = `mutation($config: UPSConfigInput!) { configureUps(config: $config) }`

const CPUSubscription = `subscription { systemMetricsCpu { percentTotal } }`
const MemorySubscription = `subscription { systemMetricsMemory { total available active percentTotal } }`
const TemperatureSubscription = `subscription { systemMetricsTemperature { sensors { name type current { value unit status } } } }`

const NetworkSubscription = `
subscription {
 systemMetricsNetwork {
  id name operstate bytesReceived bytesSent packetsReceived packetsSent receiveErrors transmitErrors
  receiveDropped transmitDropped rxSec txSec utilizationPercent lastUpdated
 }
}
`

const CPUTelemetrySubscription = `subscription { systemMetricsCpuTelemetry { totalPower power temp } }`
const DockerStatsSubscription = `subscription { dockerContainerStats { id cpuPercent memUsage memPercent netIO blockIO } }`
const LogFileSubscription = `subscription($path: String!) { logFile(path: $path) { path content totalLines startLine } }`

const NotificationSubscription = `subscription { notificationAdded { id subject importance timestamp type } }`
