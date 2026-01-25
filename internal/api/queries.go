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

const DockerContainersQuery = `
query {
	docker {
		containers {
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
