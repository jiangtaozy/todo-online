import { connect } from 'react-redux'
import { toggleTodo, getTodosIfNeeded } from '../actions'
import TodoList from '../components/TodoList'

const getVisibleTodos = (todos, filter) => {
  switch (filter) {
    case 'SHOW_COMPLETED':
      return {
        items: todos.items.filter(t => t.completed)
      }
    case 'SHOW_ACTIVE':
      return {
        items: todos.items.filter(t => !t.completed)
      }
    case 'SHOW_ALL':
    default:
      return todos
  }
}

const mapStateToProps = state => {
  return {
    todos: getVisibleTodos(state.todos, state.visibilityFilter)
  }
}

const mapDispatchToProps = dispatch => {
  return {
    onTodoClick: id => {
      dispatch(toggleTodo(id))
    },
    getTodosIfNeeded: () => dispatch(getTodosIfNeeded()), 
  }
}

const VisibleTodoList = connect(
  mapStateToProps,
  mapDispatchToProps
)(TodoList)

export default VisibleTodoList
