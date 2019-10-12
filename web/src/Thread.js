import React from 'react';
import {Link} from "react-router-dom";

import './Board.css';


class Thread extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            status: "SUCCESS",
            board: this.props.board,
            no: this.props.no,
            thread: this.props.thread,
            limit: this.props.limit
        }
    }

    componentDidMount() {
        if (this.state.thread === undefined || this.state.thread === null || this.state.status === "FAILURE" )
            this.getThread()
    }

    getThread() {
        fetch('/thread?no=' + this.state.no)
            .then((response) => {
                return response.json()
            })
            .then((json) => {
                console.log(json);
                this.setState({
                    status: json.status,
                    thread: json.thread
                })
            }).catch(console.log);
    }

    displayThread(thread) {
        return (
            <div key={thread.post.no}>
                <hr/>
                <div className="thread">
                    <span className="image"><img alt={thread.post.filename}
                                                 src={process.env.PUBLIC_URL + "/images/" + thread.post.image}/></span><span
                    className="threadHeader">{thread.subject} <span
                    className="postName">{thread.post.name}</span> {thread.post.timestamp} No. <Link
                    to={"/" + this.state.board + "/res/" + thread.post.no}>{thread.post.no}</Link></span>

                    <div><span className="content">{thread.post.comment}</span></div>
                </div>
                <div className="replies">
                    {this.displayReplies(thread, 5)}
                </div>
            </div>
        )
    }

    displayReplies(thread) {
        if (this.state.limit !== null) {
            thread.replies = thread.replies.slice(-this.state.limit)
        }
        return thread.replies.map((post) => {
            return (
                <div key={post.no} className="post">
                    {this.optionalImage(post)}
                    <span className="postHeader"><span
                        className="postName">{post.name}</span> {post.timestamp} No. {post.no}</span>
                    <div><span className="content">{post.comment}</span></div>
                </div>
            )
        })
    }

    optionalImage(post) {
        if (post.image != null && post.image !== "") {
            return <span className="image"><img src={process.env.PUBLIC_URL + "/images/" + post.image}
                                                alt={post.filename}/></span>
        } else {
            return <span className="noImage"/>
        }
    }

    render() {
        if (this.state.thread === undefined || this.state.thread === null || this.state.status === "FAILURE") {
            return (<div>. . .</div>)
        }
        return this.displayThread(this.state.thread)
    }
}

export default Thread