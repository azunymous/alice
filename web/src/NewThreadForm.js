import React from 'react';
import './Board.css';

class NewThreadForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {email: 'noko', comment: '', filename: '', image: null};
        this.fileInput = React.createRef();

        this.handleInputChange = this.handleInputChange.bind(this);
        this.handleFileChange = this.handleFileChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleInputChange(event) {
        const target = event.target;
        const value = target.type === 'checkbox' ? target.checked : target.value;
        const name = target.name;

        this.setState({
            [name]: value
        });
    }

    handleFileChange(event) {
        this.setState({image: event.target.files[0]})
    }

    handleSubmit(event) {
        event.preventDefault();
        this.uploadForm(event)
    }

    uploadForm(event) {
        let data = new FormData();

        data.append("email", this.state.email);
        data.append("comment", this.state.comment)
        data.append("image", this.state.image);
        data.append("filename", this.state.image.name)

        fetch('/thread', {
            method: 'POST',
            body: data,
        }).then((res) => {
            if (res.ok) {
                console.log("Post Success!")
                window.location.reload()
            } else {
                console.log(res.status + " " + res.statusText);
            }
        });

    }

    render() {
        return (
            <div className="reply">
                <form onSubmit={this.handleSubmit}>
                    <label>
                        Email:
                        <input type="text" name="email" value={this.state.value} onChange={this.handleInputChange}/>
                    </label>
                    <label> Comment: <textarea name="comment" value={this.state.comment}
                                               onChange={this.handleInputChange}/> </label>

                    <label htmlFor="Image">Image:</label>
                    <input type="file"
                           ref={this.fileInput}
                           onChange={this.handleFileChange}
                           id="image" name="image"
                           accept="image/png, image/jpeg"/>
                    <input type="submit" value="Submit"/>
                </form>
            </div>
        );
    }
}


export default NewThreadForm
