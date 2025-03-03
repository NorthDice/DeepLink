import React from 'react'
import './HeaderLink.css'	
import { Link } from 'react-router-dom'

const HeaderLink = ({name = "default", to = "/"}) => {
  return (
    <Link to={to} className="header__link">
        {name}
    </Link>
  )
}

export default HeaderLink
