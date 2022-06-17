import store from '@/store'
import FingerbankStoreModule from '@/views/Configuration/fingerbank/_store'
import NodesStoreModule from '@/views/Nodes/_store'

const TheView = () => import(/* webpackChunkName: "Status" */ './_components/TheView')

export default [
  {
    path: 'assets',
    name: 'assets',
    component: TheView,
    meta: {
      can: 'read nodes'
    },
    beforeEnter: (to, from, next) => {
      if (!store.state.$_fingerbank)
        store.registerModule('$_fingerbank', FingerbankStoreModule)
      if (!store.state.$_nodes)
        store.registerModule('$_nodes', NodesStoreModule)
      store.dispatch('$_fingerbank/getClasses').then(() => next())
    }
  }
]
