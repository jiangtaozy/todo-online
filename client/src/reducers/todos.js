import {
  GET_TODOS,
} from '../actions'
import { CREATE_TODO } from '../actions/createTodo'

const todos = (
  state = {
    isGetting: false,
    isCreating: false,
    items: [],
  },
  action
  ) => {
  switch (action.type) {
    case GET_TODOS:
      if(!action.status) {
        return {
          ...state,
          isGetting: true,
        }
      } else if(action.status === 'success') {
        return {
          ...state,
          isGetting: false,
          items: action.response,
        }
      } else {
        console.log('error: ', action.error)
        return {
          ...state,
          isGetting: false,
        }
      }
    case CREATE_TODO:
      if(!action.status) {
        return {
          ...state,
          isCreating: true,
        }
      } else if(action.status === 'success') {
        return {
          ...state,
          isGetting: false,
          items: [
            ...state.items,
            action.response,
          ],
        }
      } else {
        console.log('error: ', action.error)
        return {
          ...state,
          isGetting: false,
        }
      }
    case 'TOGGLE_TODO':
      return {
        items: state.items.map(todo =>
          (todo.id === action.id)
            ? {...todo, completed: !todo.completed}
            : todo
        )
      }
    default:
      return state
  }
}

export default todos
