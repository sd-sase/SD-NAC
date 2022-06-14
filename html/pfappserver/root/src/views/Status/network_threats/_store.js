/**
* "$_network_threats" store module
*/
import { createDebouncer } from 'promised-debounce'
import nodesApi from '@/views/Nodes/_api'
import securityEventsApi from './_api'

// Default values
const state = () => {
  return {
    cache: {},
    message: '',
    status: '',
    selectedCategories: [],
    selectedSecurityEvents: [],

    totalOpen: 0,
    totalClosed: 0,
    perDeviceClassOpen: false,
    perDeviceClassClosed: false,
    perSecurityEventOpen: { 'foo': 13 },
    perSecurityEventClosed: { 'foo': 12 },
  }
}

let debouncer

const getters = {
  isLoading: state => state.status === 'loading',
  perDeviceClassOpen: state => state.perDeviceClassOpen.reduce((assoc, { count = 0, device_class = '' }) => {
    return { ...assoc, [device_class]: count }
  }, {}),
  perDeviceClassClosed: state => state.perDeviceClassClosed.reduce((assoc, { count = 0, device_class = '' }) => {
    return { ...assoc, [device_class]: count }
  }, {}),
  perSecurityEventOpen: state => state.perSecurityEventOpen.reduce((assoc, { count = 0, security_event_id = '' }) => {
    return { ...assoc, [security_event_id]: count }
  }, {}),
  perSecurityEventClosed: state => state.perSecurityEventClosed.reduce((assoc, { count = 0, security_event_id = '' }) => {
    return { ...assoc, [security_event_id]: count }
  }, {}),
}

const actions = {
  stat: ({ commit }) => {
    const p1 = securityEventsApi.totalOpen().then(response => {
      const { items: [ { count = 0 } ] } = response
      commit('TOTAL_OPEN', count)
    })
    const p2 = securityEventsApi.totalClosed().then(response => {
      const { items: [ { count = 0 } ] } = response
      commit('TOTAL_CLOSED', count)
    })
    const p3 = securityEventsApi.perDeviceClassOpen().then(response => {
      commit('PER_DEVICE_CLASS_OPEN', response.items)
    })
    const p4 = securityEventsApi.perDeviceClassClosed().then(response => {
      commit('PER_DEVICE_CLASS_CLOSED', response.items)
    })
    const p5 = securityEventsApi.perSecurityEventOpen().then(response => {
      commit('PER_SECURITY_EVENT_OPEN', response.items)
    })
    const p6 = securityEventsApi.perSecurityEventClosed().then(response => {
      commit('PER_SECURITY_EVENT_CLOSED', response.items)
    })
    return Promise.all([p1, p2, p3, p4, p5, p6])
  },
  get: ({ commit }, params) => {
    let { nodes } = params
    return new Promise((resolve, reject) => {
      if (!nodes.length) {
        resolve(false)
      }
      else {
        commit('REQUEST')
        nodesApi.fingerbankCommunications({
          nodes: nodes.map(mac => mac.replace(/[^0-9A-F]/gi, ''))
        }).then(response => {
          commit('RESPONSE', response)
          resolve(true)
        }).catch(err => {
          commit('ERROR', err)
          reject(err)
        })
      }
    })
  },
  getDebounced: ({ dispatch }, params) => {
    if (!debouncer) {
      debouncer = createDebouncer()
    }
    debouncer({
      handler: () => dispatch('get', params),
      time: 100 // 100ms
    })
  },
  toggleCategory: ({ state, commit }, category) => {
    return new Promise(resolve => {
      const i = state.selectedCategories.findIndex(selected => selected === category)
      if (i > -1) {
        commit('CATEGORY_DESELECT', category)
        resolve(false)
      }
      else {
        // select category
        commit('CATEGORY_SELECT', category)
        resolve(true)
      }
    })
  },
  deselectCategories: ({ state, commit }, categories = []) => {
    return new Promise(resolve => {
      categories.forEach(category => {
        if (state.selectedCategories.indexOf(category) > -1) {
          commit('CATEGORY_DESELECT', category)
        }
      })
      resolve()
    })
  },
  selectCategories: ({ state, commit }, categories = []) => {
    return new Promise(resolve => {
      categories.forEach(category => {
        if (state.selectedCategories.indexOf(category) === -1) {
          commit('CATEGORY_SELECT', category)
        }
      })
      resolve()
    })
  },
  invertCategories: ({ state, commit }, categories = []) => {
    return new Promise(resolve => {
      categories.forEach(category => {
        if (state.selectedCategories.indexOf(category) === -1) {
          commit('CATEGORY_SELECT', category)
        }
        else {
          commit('CATEGORY_DESELECT', category)
        }
      })
      resolve()
    })
  },
  toggleSecurityEvent: ({ state, commit }, securityEvent) => {
    return new Promise(resolve => {
      const i = state.selectedSecurityEvents.findIndex(selected => selected === securityEvent)
      if (i > -1) {
        commit('SECURITY_EVENT_DESELECT', securityEvent)
        resolve(false)
      }
      else {
        commit('SECURITY_EVENT_SELECT', securityEvent)
        resolve(true)
      }
    })
  },
  deselectSecurityEvents: ({ state, commit }, securityEvents = []) => {
    return new Promise(resolve => {
      securityEvents.forEach(securityEvent => {
        if (state.selectedSecurityEvents.indexOf(securityEvent) > -1) {
          commit('SECURITY_EVENT_DESELECT', securityEvent)
        }
      })
      resolve()
    })
  },
  selectSecurityEvents: ({ state, commit }, securityEvents = []) => {
    return new Promise(resolve => {
      securityEvents.forEach(securityEvent => {
        if (state.selectedSecurityEvents.indexOf(securityEvent) === -1) {
          commit('SECURITY_EVENT_SELECT', securityEvent)
        }
      })
      resolve()
    })
  },
  invertSecurityEvents: ({ state, commit }, securityEvents = []) => {
    return new Promise(resolve => {
      securityEvents.forEach(securityEvent => {
        if (state.selectedSecurityEvents.indexOf(securityEvent) === -1) {
          commit('SECURITY_EVENT_SELECT', securityEvent)
        }
        else {
          commit('SECURITY_EVENT_DESELECT', securityEvent)
        }
      })
      resolve()
    })
  },
}

const mutations = {
  REQUEST: (state) => {
    state.status = 'loading'
    state.message = ''
  },
  RESPONSE: (state, response) => {
    state.status = 'success'
    // skip empty
    state.cache = Object.entries(response).reduce((items, [category, data]) => {
      const { all_hosts_cache = {} } = data
      if (Object.values(all_hosts_cache).length > 0) {
        items[category] = data
      }
      return items
    }, {})
  },
  ERROR: (state, response) => {
    state.status = 'error'
    if (response && response.data) {
      state.message = response.data.message
    }
  },
  CATEGORY_DESELECT: (state, category) => {
    state.selectedCategories = [ ...state.selectedCategories.filter(selected => selected !== category) ]
  },
  CATEGORY_SELECT: (state, category) => {
    state.selectedCategories.push(category)
  },
  SECURITY_EVENT_DESELECT: (state, securityEvent) => {
    state.selectedSecurityEvents = [ ...state.selectedSecurityEvents.filter(selected => selected !== securityEvent) ]
  },
  SECURITY_EVENT_SELECT: (state, securityEvent) => {
    state.selectedSecurityEvents.push(securityEvent)
  },

  TOTAL_OPEN: (state, count) => {
    state.totalOpen = count
  },
  TOTAL_CLOSED: (state, count) => {
    state.totalClosed = count
  },
  PER_DEVICE_CLASS_OPEN: (state, deviceClasses) => {
    state.perDeviceClassOpen = deviceClasses
  },
  PER_DEVICE_CLASS_CLOSED: (state, deviceClasses) => {
    state.perDeviceClassClosed = deviceClasses
  },
  PER_SECURITY_EVENT_OPEN: (state, securityEvents) => {
    state.perSecurityEventOpen = securityEvents
  },
  PER_SECURITY_EVENT_CLOSED: (state, securityEvents) => {
    state.perSecurityEventClosed = securityEvents
  }
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}