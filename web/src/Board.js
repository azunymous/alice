import React from 'react';
import './Board.css';
import Thread from "./Thread";
import NewPostForm from "./NewPostForm";

class Board extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            name: this.props.name,
            threads: []
        }
    }


    getAllThreads() {
        fetch(process.env.REACT_APP_API_URL + '/thread/all/')
            .then((response) => {
                return response.json()
            })
            .then((json) => {
                console.log(json);
                this.setState({
                    threads: json
                })
            }).catch(console.log);
    }

    componentDidMount() {
        this.getAllThreads();
    }

    render() {
        return (
            <div>
                <h2>/{this.state.name}/</h2>
                <NewPostForm/>
                {this.allThreads()}
            </div>
        );
    }

    allThreads() {
        if (this.state.threads == null) {
            return (<div> . . . </div>)
        }
        let threads = this.state.threads.slice(0, 10);
        return threads.map((thread) => {
            return (<Thread key={thread.post.no} board={this.state.name} thread={thread} limit='5'/>);
        })
    }


}

export default Board