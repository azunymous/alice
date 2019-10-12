import React from 'react';
import Board from './Board'
import './Board.css';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>alice.ws</h1>
      </header>
        <div className="outer">
          <Board name="/test/"/>
        </div>
    </div>
  );
}

export default App;
