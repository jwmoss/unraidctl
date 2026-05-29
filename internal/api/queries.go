package api

// GraphQL queries for Unraid API
// Based on actual schema from https://docs.unraid.net/API/

const InfoQuery = `
query {
	info {
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

const ArrayRemoveDiskMutation = `
mutation($input: ArrayDiskInput!) {
	array {
		removeDiskFromArray(input: $input) {
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
query {
	notifications {
		overview {
			unread {
				total
			}
		}
		list(filter: { type: UNREAD, offset: 0, limit: 50 }) {
			id
			subject
			importance
			timestamp
		}
	}
}
`

const AllNotificationsQuery = `
query {
	notifications {
		list(filter: { type: ALL, offset: 0, limit: 50 }) {
			id
			subject
			importance
			timestamp
		}
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

// Note: VMs query structure - vms.domain contains VM list
// But VMs may not be available on all systems
const VMsQuery = `
query {
	vms {
		id
		domain {
			name
			state
		}
	}
}
`
