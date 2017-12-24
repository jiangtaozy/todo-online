import axios from '../axios'

export const CREATE_TODO = 'CREATE_TODO'

// create todo
export function createTodoIfNeeded(text) {
  return (dispatch, getState) => {
    if(shouldCreateTodo(getState())) {
      return dispatch(createTodo(text))
    } else {
      return Promise.resolve()
    }
  }
}

function shouldCreateTodo(state) {
  if(state.todos.isCreating) {
    return false
  } else {
    return true
  }
}

export function createTodo(text) {
  return function(dispatch) {
    dispatch(requestCreateTodo())
    return axios.post('/todos', {
      data: {
        text: text,
      },
    }).then((response) => {
      console.log('response: ', response)
      dispatch(receiveCreateTodo(response.data.data))
    }).catch(error => {
      dispatch(receiveCreateTodo(null, error))
    })
  }
}

function requestCreateTodo() {
  return {
    type: CREATE_TODO,
  }
}

function receiveCreateTodo(result, error) {
  if(error) {
    return {
      type: CREATE_TODO,
      status: 'error',
      error: error,
    }
  } else {
    return {
      type: CREATE_TODO,
      status: 'success',
      response: result,
    }
  }
}
