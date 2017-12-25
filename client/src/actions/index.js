import axios from '../axios'

export const GET_TODOS = 'GET_TODOS'

// get todos
export function getTodosIfNeeded() {
  return (dispatch, getState) => {
    if(shouldGetTodos(getState())) {
      return dispatch(getTodos())
    } else {
      return Promise.resolve()
    }
  }
}

function shouldGetTodos(state) {
  if(state.todos.isGetting) {
    return false
  } else {
    return true
  }
}

export function getTodos() {
  return function(dispatch) {
    dispatch(requestTodos())
    return axios.get('/todos').then((response) => {
      dispatch(receiveTodos(response.data.data))
    })
  }
}

function requestTodos() {
  return {
    type: GET_TODOS
  }
}

function receiveTodos(result, error) {
  if(error) {
    return {
      type: GET_TODOS,
      status: 'error',
      error: error,
    }
  } else {
    return {
      type: GET_TODOS,
      status: 'success',
      response: result,
    }
  }
}

export const setVisibilityFilter = filter => {
  return {
    type: 'SET_VISIBILITY_FILTER',
    filter
  }
}
