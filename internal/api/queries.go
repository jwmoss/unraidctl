package api

// GraphQL queries for Unraid API

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
			threads
			speed
		}
		memory {
			total
			free
			used
		}
		versions {
			unraid
		}
	}
}
`

const ArrayStatusQuery = `
query {
	array {
		state
		capacity {
			disks {
				free
				used
				total
			}
		}
		disks {
			id
			name
			size
			status
			temp
			type
		}
		parityCheckProgress
		parityCheckRunning
	}
}
`

const ArrayStartMutation = `
mutation {
	arrayStart {
		state
	}
}
`

const ArrayStopMutation = `
mutation {
	arrayStop {
		state
	}
}
`

const DockerContainersQuery = `
query {
	dockerContainers {
		id
		names
		state
		status
		image
		autoStart
	}
}
`

const DockerStartMutation = `
mutation($id: String!) {
	dockerContainerStart(id: $id) {
		id
		state
	}
}
`

const DockerStopMutation = `
mutation($id: String!) {
	dockerContainerStop(id: $id) {
		id
		state
	}
}
`

const DockerRestartMutation = `
mutation($id: String!) {
	dockerContainerRestart(id: $id) {
		id
		state
	}
}
`

const VMsQuery = `
query {
	vms {
		id
		name
		state
		coreCount
		memory
	}
}
`

const VMStartMutation = `
mutation($id: String!) {
	vmStart(id: $id) {
		id
		state
	}
}
`

const VMStopMutation = `
mutation($id: String!) {
	vmStop(id: $id) {
		id
		state
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
		size
	}
}
`

const NotificationsQuery = `
query {
	notifications {
		id
		subject
		description
		importance
		timestamp
	}
}
`

const NotificationDismissMutation = `
mutation($id: String!) {
	notificationDismiss(id: $id)
}
`
