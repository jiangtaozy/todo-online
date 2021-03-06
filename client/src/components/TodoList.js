import React, { Component } from 'react'
import PropTypes from 'prop-types'
import Todo from './Todo'

class TodoList extends Component {
  componentDidMount() {
    const { getTodosIfNeeded } = this.props
    getTodosIfNeeded()
  }

  render() {
    let { todos, onTodoClick } = this.props
    return (
      <ul>
        {todos.items.map((todo, index) => {
          return (
            <Todo
              key={index}
              {...todo}
		          onClick={() => {
		            onTodoClick({
				          ...todo,
				          completed: !todo.completed,
	              })
	            }}
	          />
          )
        })}
      </ul>
    )
  }
}

TodoList.propTypes = {
  todos: PropTypes.shape({
    items: PropTypes.arrayOf(
      PropTypes.shape({
        id: PropTypes.string.isRequired,
        completed: PropTypes.bool.isRequired,
        text: PropTypes.string.isRequired
      }).isRequired
    ).isRequired,
  }).isRequired,
  onTodoClick: PropTypes.func.isRequired
}

export default TodoList
