import {
  GET_TODOS,
} from '../actions'
import { CREATE_TODO } from '../actions/createTodo'
import { UPDATE_TODO } from '../actions/updateTodo'

const todos = (
  state = {
    isGetting: false,
    isCreating: false,
    isUpdating: false,
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
    case UPDATE_TODO:
      if(!action.status) {
	return {
	  ...state,
	  isUpdating: true,
	}
      } else if(action.status === 'success') {
	return {
	  ...state,
	  isUpdating: false,
	  items: state.items.map(todo =>
	    (todo.id === action.response.id)
	    ? action.response : todo
	  ),
	}
      } else {
        console.log('error: ', action.error)
	return {
	  ...state,
	  isUpdating: false,
	}
      }
    default:
      return state
  }
}

export default todos
