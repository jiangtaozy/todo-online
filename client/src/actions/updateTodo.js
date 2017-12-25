import axios from '../axios'

export const UPDATE_TODO = 'UPDATE_TODO'

// update todo
export function updateTodoIfNeeded(todo) {
  return (dispatch, getState) => {
    if(shouldCreateTodo(getState())) {
      return dispatch(updateTodo(todo))
    } else {
      return Promise.resolve()
    }
  }
}

function shouldCreateTodo(state) {
  if(state.todos.isUpdating) {
    return false
  } else {
    return true
  }
}

export function updateTodo(todo) {
  return function(dispatch) {
    dispatch(requestUpdateTodo())
    return axios.put('/todos/' + todo.id, {
      data: {
        completed: todo.completed,
        text: todo.text,
      }
    }).then(response => {
      dispatch(receiveUpdateTodo(todo))
    }).catch(error => {
      dispatch(receiveUpdateTodo(null, error))
    })
  }
}

function requestUpdateTodo() {
  return {
    type: UPDATE_TODO,
  }
}

function receiveUpdateTodo(result, error) {
  if(error) {
    return {
      type: UPDATE_TODO,
      status: 'error',
      error: error,
    }
  } else {
    return {
      type: UPDATE_TODO,
      status: 'success',
      response: result,
    }
  }
}
