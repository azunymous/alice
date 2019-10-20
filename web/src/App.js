import React from 'react';
import {BrowserRouter as Router, Link, Route, Switch, useParams} from "react-router-dom";

import Board from './Board'
import Thread from "./Thread";
import './Board.css';
import NewPostForm from "./NewPostForm";

function App() {
    return (
        <div className="App">
            <header className="App-header">
                <h1>igiari.net</h1>
            </header>
            <div className="outer">
                <Router>
                    <span><ul id="menu"><li><Link to="/">Home</Link></li> <li><Link
                        to="/obj/">/obj/</Link></li></ul></span>
                    <Switch>
                        <Route exact path="/">
                            . . .
                        </Route>
                        <Route exact path="/:boardID/" children={<ShowBoard/>}/>
                        <Route path="/:boardID/res/:threadNo" children={<ShowThread/>}/>
                    </Switch>
                </Router>
            </div>
        </div>
    );

    function ShowBoard() {
        let {boardID} = useParams();
        return (
            <Board name={boardID}/>
        );
    }

    function ShowThread() {
        let {boardID, threadNo} = useParams();
        return (
            <div>
                <NewPostForm board={boardID} threadNo={threadNo}/>
                <Thread board={boardID} no={threadNo}/>
            </div>
        );
    }
}


export default App;
