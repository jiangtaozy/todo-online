import React from 'react'
import PropTypes from 'prop-types'

const Todo = ({ onClick, completed, text, create_at }) => {
  let createAt = new Date(create_at)
  return (
    <li
      onClick={onClick}
      style={{
        textDecoration: completed ? 'line-through' : 'none',
        padding: 10,
        display: 'flex',
      }}
    >
      <div style={{
        display: 'flex',
        alignItems: 'center',
      }}>
        <div>
          {text}
        </div>
        <div style={{
          fontSize: 12,
          color: 'gray',
          paddingLeft: 10,
        }}>
          {createAt.toLocaleDateString('zh-Hans-CN')}
        </div>
      </div>
    </li>
  )
}

Todo.propTypes = {
  onClick: PropTypes.func.isRequired,
  completed: PropTypes.bool.isRequired,
  text: PropTypes.string.isRequired
}

export default Todo
