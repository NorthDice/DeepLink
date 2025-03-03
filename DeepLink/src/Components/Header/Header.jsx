import React from "react"
import "./Header.css"
import "../HeaderLink/HeaderLink.jsx"
import { Link } from "react-router-dom";
import HeaderLink from "../HeaderLink/HeaderLink.jsx";

const Header = () => {
  return (
    <header className="header">

        <div className="header-logo">
            <Link to="/" className="header-logo__link">DeepLink</Link>
        </div>

        <nav className="header-nav">
            <ul className="header-nav__list">
                <li className="header-nav__item">
                    <HeaderLink name="Home"/>
                </li>
                <li className="header-nav__item">
                    <HeaderLink name="Notifications"/>
                </li>
                <li className="header-nav__item">
                    <HeaderLink name="Messages"/>
                </li>
                <li className="header-nav__item">
                    <HeaderLink name="Profile"/>
                </li>
            </ul>
        </nav>

        <div className="header-post">
            <button className="header-post__button">Post</button>
        </div>

        <div className="header-user">
            <button className="header-user__button">
                <p className="button-paragrahp">Username</p>
                <p className="button-paragraph">Tag</p>
            </button>
        </div>

    </header>
  )
}

export default Header
