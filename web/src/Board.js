import React from 'react';
import './style.css';

class Board extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            threads: []
        }
    }


    getAllThreads() {
        fetch('/thread/all/')
            .then((response) => {
                return response.json()
            })
            .then((json) => {
                console.log(json)
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
                <h2>Hello, {this.props.name}</h2>
                {this.allThreads()}
            </div>
        );
    }

    allThreads() {
        if (this.state.threads == null) {
            return (<div>Loading...</div>)
        }
        return this.state.threads.map((thread) => {
            return (
                <div key={thread.post.no}>
                    <div className="thread">
                        <span className="threadHeader">{thread.subject} <span className="postName">{thread.post.name}</span> {thread.post.timestamp} No. {thread.post.no}</span>
                        <div>{thread.post.comment}</div>
                    </div>
                    <div className="replies">
                        {this.replies(thread, 5)}
                    </div>
                </div>
            )
        })
    }

    replies(thread, limit = null) {
        if (limit !== null) {
            thread.replies = thread.replies.slice(-limit)
        }
        return thread.replies.map((post) => {
            return (
                <div key={post.no} className="post">
                    <div className="postHeader"><span className="postName">{post.name}</span> {post.timestamp} No. {post.no}</div>
                    <div>{post.comment}</div>
                </div>
            )
        })
    }
}

export default Board